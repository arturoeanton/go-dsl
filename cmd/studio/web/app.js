(() => {
  const state = {
    presets: [],
    filePath: "",
    config: null,
    gbnf: "",
    promptSpec: "",
    specView: "gbnf",
    aiProvider: "ollama",
    aiMode: "generate_code",
    aiStatus: null,
    lastAIResult: null,
  };

  const $ = (id) => document.getElementById(id);

  const grammarEditor = $("grammarEditor");
  const codeEditor = $("codeEditor");
  const grammarGutter = $("grammarGutter");
  const codeGutter = $("codeGutter");

  function showToast(msg) {
    const t = $("toast");
    t.textContent = msg;
    t.classList.remove("hidden");
    clearTimeout(showToast._timer);
    showToast._timer = setTimeout(() => t.classList.add("hidden"), 2800);
  }

  function updateGutter(textarea, gutter) {
    const lines = textarea.value.split("\n").length;
    let out = "";
    for (let i = 1; i <= lines; i++) {
      out += i + (i < lines ? "\n" : "");
    }
    gutter.textContent = out;
  }

  function syncScroll(textarea, gutter) {
    gutter.scrollTop = textarea.scrollTop;
  }

  function updateCursorLabel() {
    const pos = codeEditor.selectionStart || 0;
    const before = codeEditor.value.slice(0, pos);
    const lines = before.split("\n");
    const ln = lines.length;
    const col = lines[lines.length - 1].length + 1;
    $("cursorPosLabel").textContent = `Ln ${ln}, Col ${col}`;
    fetchCompletionsDebounced();
  }

  function debounce(fn, ms) {
    let t;
    return (...args) => {
      clearTimeout(t);
      t = setTimeout(() => fn(...args), ms);
    };
  }

  async function postJSON(url, body) {
    const resp = await fetch(url, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body || {}),
    });
    return resp.json();
  }

  // --- Initialization ---
  async function initStudio() {
    const res = await fetch("/api/studio/init").then((r) => r.json());
    state.presets = res.presets || [];
    state.filePath = res.file_path || "";

    const sel = $("presetSelect");
    sel.innerHTML = "";
    state.presets.forEach((p, idx) => {
      const opt = document.createElement("option");
      opt.value = p.id;
      opt.textContent = `${p.title} (${p.badge})`;
      if (idx === 0) opt.selected = true;
      sel.appendChild(opt);
    });

    if (state.filePath) {
      $("filePathBadge").textContent = state.filePath.split("/").pop();
    }

    grammarEditor.value = res.grammar_yaml || "";
    codeEditor.value = res.sample_code || "";
    updateGutter(grammarEditor, grammarGutter);
    updateGutter(codeEditor, codeGutter);

    if (res.ai_status) {
      applyAIStatus(res.ai_status);
    }

    await reloadGrammarAndEval();
  }

  // --- Grammar Loading & Visual Map ---
  async function reloadGrammarAndEval() {
    const loadRes = await postJSON("/api/studio/load", {
      grammar: grammarEditor.value,
      format: "auto",
    });

    state.config = loadRes.config || null;
    state.gbnf = loadRes.gbnf || "";
    state.promptSpec = loadRes.prompt_spec || "";

    renderGrammarHealth(loadRes);
    renderVisualMap(state.config);
    renderSpecViewer();
    refreshCodegenPreview();

    if (loadRes.ok) {
      await evaluateCode();
      await fetchCompletions();
    }
  }

  function renderGrammarHealth(loadRes) {
    const pill = $("grammarStatusPill");
    const txt = $("grammarStatusText");
    const tray = $("grammarIssuesBar");
    pill.className = "status-pill";

    // Filter out expected "unregistered action" informational warnings in Studio mode
    const structuralWarnings = (loadRes.warnings || []).filter(
      (w) => !w.includes("unregistered action")
    );

    if (!loadRes.ok) {
      pill.classList.add("err");
      txt.textContent = "Error en gramática";
      tray.classList.remove("hidden");
      tray.innerHTML = `<div class="diag-title">Error estructural en gramática</div><div class="diag-item">${escapeHTML(loadRes.error || "")}</div>`;
    } else if (structuralWarnings.length > 0) {
      pill.classList.add("warn");
      txt.textContent = `Válida (${structuralWarnings.length} avisos)`;
      tray.classList.remove("hidden");
      tray.innerHTML = structuralWarnings
        .map((w) => `<div class="diag-item">⚠ ${escapeHTML(w)}</div>`)
        .join("");
    } else {
      pill.classList.add("ok");
      const name = state.config?.name || "DSL";
      txt.textContent = `${name} · Gramática OK`;
      tray.classList.add("hidden");
      tray.innerHTML = "";
    }
  }

  function renderVisualMap(cfg) {
    const box = $("grammarVisualView");
    if (!cfg) {
      box.innerHTML = `<p class="panel-hint">Corrige el YAML para visualizar la arquitectura de la gramática.</p>`;
      return;
    }

    const tokHTML = (cfg.token_defs || [])
      .map((t) => {
        const k = (t.kind || "regex").toLowerCase();
        return `<span class="tok-badge ${escapeHTML(k)}" title="${escapeHTML(t.pattern)}">${escapeHTML(t.name)} <small>(${escapeHTML(k)})</small></span>`;
      })
      .join("");

    const rulesHTML = (cfg.rules || [])
      .map((r) => {
        const syms = (r.pattern || [])
          .map((s) => `<span class="rule-sym">${escapeHTML(s)}</span>`)
          .join("");
        const act = r.action ? `<span class="rule-act">→ ${escapeHTML(r.action)}</span>` : "";
        return `<div class="rule-diagram"><span class="rule-lhs">${escapeHTML(r.name)}</span><span class="rule-arrow">::=</span>${syms}${act}</div>`;
      })
      .join("");

    const exprHTML = (cfg.expressions || [])
      .map((ex) => {
        const rows = [];
        (ex.postfix || []).forEach((op) => rows.push({ p: op.power, desc: `Postfix ${op.token}`, act: op.action }));
        (ex.calls || []).forEach((c) => rows.push({ p: c.power, desc: `Call ${c.open}...${c.close}`, act: c.action }));
        (ex.members || []).forEach((m) => rows.push({ p: m.power, desc: `Member ${m.dot_token}${m.field_token}`, act: m.action }));
        (ex.prefix || []).forEach((op) => rows.push({ p: op.power, desc: `Prefix ${op.token}`, act: op.action }));
        (ex.infix || []).forEach((op) => rows.push({ p: op.power, desc: `Infix (${op.assoc || "left"}) ${op.token}`, act: op.action }));
        if (ex.ternary) {
          rows.push({ p: ex.ternary.power, desc: `Ternary ${ex.ternary.question_token} : ${ex.ternary.colon_token}`, act: ex.ternary.action });
        }
        rows.sort((a, b) => b.p - a.p);

        const ladder = rows
          .map(
            (r) =>
              `<div class="rule-diagram"><span class="rule-lhs">bp ${r.p}</span><span class="rule-sym">${escapeHTML(r.desc)}</span><span class="rule-act">${escapeHTML(r.act || "")}</span></div>`
          )
          .join("");

        return `<div class="vmap-section"><div class="vmap-title">Escalera Pratt · ${escapeHTML(ex.name)}</div>${ladder || '<span class="panel-hint">Solo átomos</span>'}</div>`;
      })
      .join("");

    box.innerHTML = `
      <div class="vmap-section">
        <div class="vmap-title">Léxico Ordenado (${(cfg.token_defs || []).length} tokens) · Start: ${escapeHTML(cfg.start_rule || "auto")}</div>
        <div class="chip-cloud">${tokHTML || "—"}</div>
      </div>
      ${rulesHTML ? `<div class="vmap-section"><div class="vmap-title">Producciones BNF / PEG</div>${rulesHTML}</div>` : ""}
      ${exprHTML}
    `;
  }

  // --- Evaluation, AST, Trace, Tokens & Diagnostics ---
  async function evaluateCode() {
    const maxSteps = parseInt($("maxStepsInput").value, 10) || 10000;
    const res = await postJSON("/api/studio/eval", {
      grammar: grammarEditor.value,
      code: codeEditor.value,
      max_steps: maxSteps,
    });

    renderTokenStream(res.tokens || [], res.trivia || []);
    renderDiagnostics(res.diagnostics || [], res.ok ? "" : res.error);

    if (res.ok) {
      $("outputValue").textContent = formatOutputValue(res.output);
      $("astPrettyBox").textContent = res.ast_pretty || "";
      renderASTTree(res.ast);
    } else {
      $("outputValue").textContent = "✗ Error de parseo / evaluación";
      renderASTTree(null);
    }

    renderTrace(res.trace);
  }

  function formatOutputValue(val) {
    if (val === undefined || val === null) return "nil";
    if (typeof val === "object") return JSON.stringify(val, null, 2);
    return String(val);
  }

  function renderTokenStream(tokens, trivia) {
    const all = [
      ...tokens.map((t) => ({ ...t, isTrivia: false })),
      ...trivia.map((t) => ({ ...t, isTrivia: true })),
    ].sort((a, b) => a.Start - b.Start);

    $("tokenCountLabel").textContent = String(tokens.length);
    const ribbon = $("tokenRibbon");
    ribbon.innerHTML = "";

    all.forEach((tok) => {
      const el = document.createElement("div");
      el.className = "stream-tok" + (tok.isTrivia ? " trivia" : "");
      el.innerHTML = `<span class="tt">${escapeHTML(tok.TokenType)}</span><span>${escapeHTML(JSON.stringify(tok.Value))}</span>`;
      el.addEventListener("click", () => highlightSourceSpan(tok.Start, tok.End));
      ribbon.appendChild(el);
    });
  }

  function renderDiagnostics(diags, fallbackErr) {
    const panel = $("diagnosticsPanel");
    const list = $("diagList");
    if ((!diags || diags.length === 0) && !fallbackErr) {
      panel.classList.add("hidden");
      list.innerHTML = "";
      return;
    }

    panel.classList.remove("hidden");
    if (diags && diags.length > 0) {
      $("diagSummary").textContent = `${diags.length} diagnóstico(s) detectado(s)`;
      list.innerHTML = diags
        .map((d) => `<div class="diag-item">${escapeHTML(d.detailed || `Ln ${d.line}, Col ${d.column}: ${d.message}`)}</div>`)
        .join("");
    } else {
      $("diagSummary").textContent = "Error de ejecución";
      list.innerHTML = `<div class="diag-item">${escapeHTML(fallbackErr)}</div>`;
    }
  }

  function renderASTTree(ast) {
    const container = $("astTree");
    container.innerHTML = "";
    if (!ast) {
      container.innerHTML = `<span class="panel-hint">Sin AST disponible (corrige errores sintácticos).</span>`;
      return;
    }
    container.appendChild(buildASTDOM(ast));
  }

  function buildASTDOM(node) {
    const wrap = document.createElement("div");
    const row = document.createElement("div");
    row.className = "tree-node";

    if (node.token) {
      row.innerHTML = `<span class="badge-tok">● ${escapeHTML(node.token)}</span><span class="badge-val">${escapeHTML(JSON.stringify(node.value || ""))}</span><span class="badge-span">[${node.start}..${node.end}]</span>`;
    } else {
      const act = node.action ? `<span class="badge-act">→ ${escapeHTML(node.action)}</span>` : "";
      row.innerHTML = `<span class="badge-rule">◆ ${escapeHTML(node.rule)}</span>${act}<span class="badge-span">[${node.start}..${node.end}]</span>`;
    }
    row.addEventListener("click", (e) => {
      e.stopPropagation();
      highlightSourceSpan(node.start, node.end);
    });
    wrap.appendChild(row);

    if (node.children && node.children.length > 0) {
      const chBox = document.createElement("div");
      chBox.className = "tree-children";
      node.children.forEach((c) => chBox.appendChild(buildASTDOM(c)));
      wrap.appendChild(chBox);
    }
    return wrap;
  }

  function renderTrace(trace) {
    const box = $("traceTree");
    box.innerHTML = "";
    if (!trace || !trace.root) {
      $("traceMetrics").textContent = "0 pasos";
      box.innerHTML = `<span class="panel-hint">Ejecuta un script válido para inspeccionar la traza de evaluación.</span>`;
      return;
    }
    const durUs = Math.max(1, Math.round((trace.total_duration_ns || 0) / 1000));
    $("traceMetrics").textContent = `${trace.steps || 0} pasos · ${durUs} µs`;
    box.appendChild(buildTraceDOM(trace.root));
  }

  function buildTraceDOM(tn) {
    const wrap = document.createElement("div");
    const row = document.createElement("div");
    row.className = "tree-node";

    const skipBadge = tn.skipped ? `<span class="badge-skip">[SKIPPED]</span>` : "";
    const label = tn.rule
      ? `<span class="badge-rule">${escapeHTML(tn.rule)}</span>`
      : `<span class="badge-tok">${escapeHTML(tn.token || "tok")}</span>`;
    const act = tn.action ? `<span class="badge-act">(${escapeHTML(tn.action)})</span>` : "";
    const outVal = !tn.skipped && tn.output !== undefined
      ? `<span class="badge-val">⇒ ${escapeHTML(JSON.stringify(tn.output))}</span>`
      : "";
    const dur = tn.duration_ns ? `<span class="badge-span">${Math.max(1, Math.round(tn.duration_ns / 1000))}µs</span>` : "";

    row.innerHTML = `${skipBadge}${label}${act}${outVal}${dur}`;
    row.addEventListener("click", () => {
      if (typeof tn.span_start === "number" && typeof tn.span_end === "number") {
        highlightSourceSpan(tn.span_start, tn.span_end);
      }
    });
    wrap.appendChild(row);

    if (tn.children && tn.children.length > 0) {
      const ch = document.createElement("div");
      ch.className = "tree-children";
      tn.children.forEach((c) => ch.appendChild(buildTraceDOM(c)));
      wrap.appendChild(ch);
    }
    return wrap;
  }

  function highlightSourceSpan(start, end) {
    const text = codeEditor.value;
    const byteToCharIndex = (byteOffset) => {
      const enc = new TextEncoder();
      let bytes = 0;
      for (let i = 0; i < text.length; i++) {
        if (bytes >= byteOffset) return i;
        bytes += enc.encode(text[i]).length;
      }
      return text.length;
    };
    const sIdx = byteToCharIndex(start);
    const eIdx = byteToCharIndex(end);
    codeEditor.focus();
    codeEditor.setSelectionRange(sIdx, eIdx);
    updateCursorLabel();
  }

  // --- Parser-driven Completions ---
  async function fetchCompletions() {
    const charPos = codeEditor.selectionStart || 0;
    const byteOffset = new TextEncoder().encode(codeEditor.value.slice(0, charPos)).length;
    const res = await postJSON("/api/studio/complete", {
      code: codeEditor.value,
      offset: byteOffset,
    });
    const list = $("completionsList");
    list.innerHTML = "";
    const comps = res.completions || [];
    if (comps.length === 0) {
      list.innerHTML = `<span class="panel-hint">fin de sentencia / sin sugerencias</span>`;
      return;
    }
    comps.forEach((c) => {
      const btn = document.createElement("button");
      btn.className = "comp-chip" + (c.IsKeyword ? "" : " freeform");
      btn.textContent = c.Label;
      btn.title = c.Detail || "";
      btn.addEventListener("click", () => insertAtCursor(c.IsKeyword ? c.Label + " " : ""));
      list.appendChild(btn);
    });
  }
  const fetchCompletionsDebounced = debounce(fetchCompletions, 120);

  function insertAtCursor(snippet) {
    if (!snippet) return;
    const start = codeEditor.selectionStart || 0;
    const end = codeEditor.selectionEnd || 0;
    const v = codeEditor.value;
    codeEditor.value = v.slice(0, start) + snippet + v.slice(end);
    codeEditor.focus();
    codeEditor.setSelectionRange(start + snippet.length, start + snippet.length);
    updateGutter(codeEditor, codeGutter);
    evaluateCode();
    fetchCompletions();
  }

  // --- Local AI Copilot (Ollama & LM Studio) ---
  async function refreshAIStatus() {
    $("ollamaCardStatus").textContent = "escaneando...";
    $("lmstudioCardStatus").textContent = "escaneando...";
    const status = await postJSON("/api/ai/status", {
      ollama_url: $("ollamaUrlInput").value,
      lmstudio_url: $("lmstudioUrlInput").value,
    });
    applyAIStatus(status);
  }

  function applyAIStatus(status) {
    state.aiStatus = status;
    const ol = status.ollama || {};
    const lm = status.lmstudio || {};

    updateProviderPill("ollamaBadge", "ollamaCardStatus", "ollamaCount", ol);
    updateProviderPill("lmstudioBadge", "lmstudioCardStatus", "lmstudioCount", lm);

    // Auto-select online provider if current one is offline
    if (state.aiProvider === "ollama" && !ol.online && lm.online) {
      state.aiProvider = "lmstudio";
    } else if (state.aiProvider === "lmstudio" && !lm.online && ol.online) {
      state.aiProvider = "ollama";
    }
    syncProviderSelectionUI();
  }

  function updateProviderPill(badgeId, cardStatusId, countId, info) {
    const badge = $(badgeId);
    const cardSt = $(cardStatusId);
    const cnt = $(countId);

    if (info.online) {
      badge.classList.add("online");
      cnt.textContent = `${(info.models || []).length}`;
      cardSt.textContent = `Online · ${info.latency_ms}ms`;
      cardSt.className = "mini-status online";
    } else {
      badge.classList.remove("online");
      cnt.textContent = "off";
      cardSt.textContent = "Offline";
      cardSt.className = "mini-status";
    }
  }

  function syncProviderSelectionUI() {
    $("cardOllama").classList.toggle("active", state.aiProvider === "ollama");
    $("cardLMStudio").classList.toggle("active", state.aiProvider === "lmstudio");

    const info = state.aiStatus ? state.aiStatus[state.aiProvider] : null;
    const sel = $("aiModelSelect");
    sel.innerHTML = "";

    const hintBox = $("aiSetupHint");
    if (!info || !info.online || !(info.models || []).length) {
      const opt = document.createElement("option");
      opt.value = "";
      opt.textContent = info?.online ? "Sin modelos cargados" : `${state.aiProvider} no detectado en localhost`;
      sel.appendChild(opt);
      hintBox.textContent = info?.setup_hint || "Inicia Ollama o LM Studio localmente y pulsa ↻ Detectar.";
      hintBox.classList.remove("hidden");
      return;
    }

    hintBox.classList.add("hidden");
    info.models.forEach((m) => {
      const opt = document.createElement("option");
      opt.value = m.id;
      const meta = [m.parameter_size, m.quantization].filter(Boolean).join(" · ");
      opt.textContent = (m.recommended ? "★ " : "") + m.id + (meta ? ` (${meta})` : "");
      if (m.id === info.recommended) opt.selected = true;
      sel.appendChild(opt);
    });
  }

  async function runAICopilot() {
    const btn = $("btnRunAI");
    const origLabel = btn.textContent;
    btn.disabled = true;
    btn.textContent = "⏳ Generando y verificando con go-dsl...";

    try {
      const baseUrl = state.aiProvider === "ollama" ? $("ollamaUrlInput").value : $("lmstudioUrlInput").value;
      const res = await postJSON("/api/ai/run", {
        provider: state.aiProvider,
        base_url: baseUrl,
        model: $("aiModelSelect").value,
        mode: state.aiMode,
        prompt: $("aiPromptInput").value,
        code: codeEditor.value,
        grammar_yaml: grammarEditor.value,
        use_gbnf: $("chkUseGBNF").checked,
        temperature: 0.15,
      });

      state.lastAIResult = res;
      const card = $("aiResultCard");
      card.classList.remove("hidden");

      if (!res.ok) {
        $("aiVerifyBadge").textContent = "Error";
        $("aiVerifyBadge").className = "verify-pill unverified";
        $("aiResultPreview").textContent = res.error || "Error desconocido";
        $("aiTelemetryMeta").textContent = "";
        $("aiGuardrailLog").innerHTML = "";
        return;
      }

      const vBadge = $("aiVerifyBadge");
      if (res.verified) {
        vBadge.textContent = "✓ Verificado por go-dsl";
        vBadge.className = "verify-pill";
      } else {
        vBadge.textContent = "⚠ Requiere revisión";
        vBadge.className = "verify-pill unverified";
      }

      const gbnfTag = res.gbnf_used ? "GBNF activo" : res.gbnf_fallback ? "Fallback RepairPrompt" : "PromptSpec";
      $("aiTelemetryMeta").textContent = `${res.attempts} intento(s) · ${gbnfTag} · ${res.duration_ms}ms`;
      $("aiResultPreview").textContent = res.result_code || res.result_yaml || "";
      $("aiGuardrailLog").innerHTML = (res.log || [])
        .map((l) => `<div>• ${escapeHTML(l)}</div>`)
        .join("");
    } finally {
      btn.disabled = false;
      btn.textContent = origLabel;
    }
  }

  async function applyAIResult() {
    const r = state.lastAIResult;
    if (!r || !r.ok) return;

    if (r.result_yaml) {
      grammarEditor.value = r.result_yaml;
      updateGutter(grammarEditor, grammarGutter);
      if (r.sample_code) {
        codeEditor.value = r.sample_code;
        updateGutter(codeEditor, codeGutter);
      }
      await reloadGrammarAndEval();
      showToast("✓ Gramática aplicada y validada en vivo");
    } else if (r.result_code) {
      codeEditor.value = r.result_code;
      updateGutter(codeEditor, codeGutter);
      await evaluateCode();
      showToast("✓ Código DSL aplicado y evaluado");
    }
  }

  // --- Tests, Fuzzing, GBNF & Export ---
  async function runSuiteAndOptionalFuzz(fuzzIters) {
    const res = await postJSON("/api/studio/tests", {
      fuzz_iterations: fuzzIters || 0,
      seed: Date.now() % 100000,
    });

    const fuzzBox = $("fuzzSummaryBox");
    if (res.fuzz_report) {
      const fr = res.fuzz_report;
      fuzzBox.classList.remove("hidden");
      fuzzBox.textContent = `✓ Fuzzer: ${fr.iterations} mutaciones ejecutadas · ${fr.valid_parsed} válidas · ${fr.rejected_cleanly} rechazadas limpiamente · ${(fr.failures || []).length} panics`;
    } else {
      fuzzBox.classList.add("hidden");
    }

    const list = $("testResultsList");
    const tr = res.test_report;
    if (!tr || !tr.results || tr.results.length === 0) {
      list.innerHTML = `<p class="panel-hint">No hay casos en \`tests:\` dentro del YAML. Usa la pestaña 🤖 IA Local → Generar Tests.</p>`;
      return;
    }
    list.innerHTML = tr.results
      .map(
        (r) => `
        <div class="test-card ${r.passed ? "pass" : "fail"}">
          <strong>${r.passed ? "✓" : "✗"} ${escapeHTML(r.name)}</strong>
          ${r.message ? `<div class="panel-hint">${escapeHTML(r.message)}</div>` : ""}
        </div>`
      )
      .join("");
  }

  function renderSpecViewer() {
    $("specViewer").textContent = state.specView === "gbnf" ? state.gbnf : state.promptSpec;
  }

  async function refreshCodegenPreview() {
    const res = await postJSON("/api/studio/codegen", {
      package_name: $("pkgNameInput").value || "mydsl",
      func_name: $("funcNameInput").value || "NewMyDSL",
      typed: $("chkTypedCodegen").checked,
    });
    $("codegenPreview").textContent = res.ok ? res.go_source : `// Error: ${res.error}`;
  }

  function downloadBlob(filename, content, mime) {
    const blob = new Blob([content], { type: mime || "text/plain;charset=utf-8" });
    const a = document.createElement("a");
    a.href = URL.createObjectURL(blob);
    a.download = filename;
    a.click();
    URL.revokeObjectURL(a.href);
  }

  function switchInspectorTab(tabName) {
    document.querySelectorAll(".itab").forEach((b) => b.classList.toggle("active", b.dataset.tab === tabName));
    document.querySelectorAll(".tab-body").forEach((p) => p.classList.toggle("active", p.id === `tab-${tabName}`));
    if (tabName === "tests") runSuiteAndOptionalFuzz(0);
    if (tabName === "export") refreshCodegenPreview();
  }

  function escapeHTML(s) {
    return String(s)
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;");
  }

  // --- Event Listeners ---
  $("presetSelect").addEventListener("change", async (e) => {
    const p = state.presets.find((x) => x.id === e.target.value);
    if (!p) return;
    grammarEditor.value = p.grammar_yaml;
    codeEditor.value = p.sample_code;
    updateGutter(grammarEditor, grammarGutter);
    updateGutter(codeEditor, codeGutter);
    await reloadGrammarAndEval();
  });

  const debouncedReloadGrammar = debounce(reloadGrammarAndEval, 320);
  const debouncedEvalCode = debounce(evaluateCode, 180);

  grammarEditor.addEventListener("input", () => {
    updateGutter(grammarEditor, grammarGutter);
    debouncedReloadGrammar();
  });
  grammarEditor.addEventListener("scroll", () => syncScroll(grammarEditor, grammarGutter));

  codeEditor.addEventListener("input", () => {
    updateGutter(codeEditor, codeGutter);
    updateCursorLabel();
    debouncedEvalCode();
  });
  codeEditor.addEventListener("click", updateCursorLabel);
  codeEditor.addEventListener("keyup", updateCursorLabel);
  codeEditor.addEventListener("scroll", () => syncScroll(codeEditor, codeGutter));

  // Tab indentation in editors + Cmd/Ctrl+Enter shortcut
  [grammarEditor, codeEditor].forEach((ed) => {
    ed.addEventListener("keydown", (e) => {
      if (e.key === "Tab") {
        e.preventDefault();
        const s = ed.selectionStart;
        const end = ed.selectionEnd;
        ed.value = ed.value.slice(0, s) + "  " + ed.value.slice(end);
        ed.selectionStart = ed.selectionEnd = s + 2;
        ed.dispatchEvent(new Event("input"));
      }
    });
  });

  window.addEventListener("keydown", (e) => {
    if ((e.metaKey || e.ctrlKey) && e.key === "Enter") {
      e.preventDefault();
      evaluateCode();
    } else if (e.shiftKey && e.altKey && e.key.toLowerCase() === "f") {
      e.preventDefault();
      $("btnFormat").click();
    }
  });

  $("btnRun").addEventListener("click", evaluateCode);

  $("btnFormat").addEventListener("click", async () => {
    const res = await postJSON("/api/studio/format", { code: codeEditor.value });
    if (res.ok && res.formatted !== undefined) {
      codeEditor.value = res.formatted;
      updateGutter(codeEditor, codeGutter);
      await evaluateCode();
      showToast("✨ Código formateado canónicamente");
    } else {
      showToast("✗ Corrige los errores sintácticos antes de formatear");
    }
  });

  $("btnSample").addEventListener("click", async () => {
    const res = await postJSON("/api/studio/tests", { sample_count: 1, seed: Date.now() });
    if (res.samples && res.samples.length > 0) {
      codeEditor.value = res.samples[0];
      updateGutter(codeEditor, codeGutter);
      await evaluateCode();
      showToast("🎲 Ejemplo válido sintetizado desde la gramática");
    }
  });

  $("btnSaveGrammar").addEventListener("click", async () => {
    if (state.filePath) {
      const res = await postJSON("/api/studio/save", { grammar: grammarEditor.value });
      if (res.ok) {
        showToast(`💾 Guardado en ${res.saved_to}`);
        return;
      }
    }
    const slug = (state.config?.name || "grammar").toLowerCase().replace(/[^a-z0-9]+/g, "-");
    downloadBlob(`${slug}.yaml`, grammarEditor.value, "text/yaml;charset=utf-8");
    showToast(`⬇ Descargado ${slug}.yaml`);
  });

  // Grammar YAML vs Visual Map toggle
  document.querySelectorAll("[data-gview]").forEach((btn) => {
    btn.addEventListener("click", () => {
      document.querySelectorAll("[data-gview]").forEach((b) => b.classList.toggle("active", b === btn));
      const isVis = btn.dataset.gview === "visual";
      $("grammarYamlView").classList.toggle("hidden", isVis);
      $("grammarVisualView").classList.toggle("hidden", !isVis);
    });
  });

  // Inspector Tabs
  document.querySelectorAll(".itab").forEach((btn) => {
    btn.addEventListener("click", () => switchInspectorTab(btn.dataset.tab));
  });

  $("btnTogglePrettyAST").addEventListener("click", () => {
    $("astPrettyBox").classList.toggle("hidden");
  });

  // Quick Repair from Diagnostics Banner
  $("btnQuickRepair").addEventListener("click", () => {
    switchInspectorTab("ai");
    document.querySelector('[data-aimode="repair_code"]').click();
    runAICopilot();
  });

  // AI Radar badges open AI tab
  $("ollamaBadge").addEventListener("click", () => {
    state.aiProvider = "ollama";
    syncProviderSelectionUI();
    switchInspectorTab("ai");
  });
  $("lmstudioBadge").addEventListener("click", () => {
    state.aiProvider = "lmstudio";
    syncProviderSelectionUI();
    switchInspectorTab("ai");
  });

  $("cardOllama").addEventListener("click", () => {
    state.aiProvider = "ollama";
    syncProviderSelectionUI();
  });
  $("cardLMStudio").addEventListener("click", () => {
    state.aiProvider = "lmstudio";
    syncProviderSelectionUI();
  });
  $("btnRefreshAI").addEventListener("click", refreshAIStatus);

  document.querySelectorAll(".mode-pill").forEach((pill) => {
    pill.addEventListener("click", () => {
      document.querySelectorAll(".mode-pill").forEach((p) => p.classList.toggle("active", p === pill));
      state.aiMode = pill.dataset.aimode;
      const placeholders = {
        generate_code: "Describe la regla de negocio que quieres escribir en tu DSL...",
        repair_code: "Opcional: añade contexto extra sobre lo que intentabas expresar (o déjalo vacío para reparación automática)...",
        generate_grammar: "Describe el nuevo lenguaje DSL que quieres crear (ej: DSL para alertas IoT con umbrales de temperatura, comentarios # y operadores booleanos)...",
        generate_tests: "Enfoque de dominio opcional para los casos de prueba (ej: valores límite, montos cero, casos inválidos)...",
      };
      $("aiPromptInput").placeholder = placeholders[state.aiMode] || "";
    });
  });

  $("btnRunAI").addEventListener("click", runAICopilot);
  $("btnApplyAIResult").addEventListener("click", applyAIResult);

  // Tests & Fuzzer
  $("btnRunTests").addEventListener("click", () => runSuiteAndOptionalFuzz(0));
  $("btnRunFuzz").addEventListener("click", () => runSuiteAndOptionalFuzz(300));

  // GBNF / Prompt Spec toggle
  document.querySelectorAll("[data-specview]").forEach((btn) => {
    btn.addEventListener("click", () => {
      document.querySelectorAll("[data-specview]").forEach((b) => b.classList.toggle("active", b === btn));
      state.specView = btn.dataset.specview;
      renderSpecViewer();
    });
  });

  $("btnCopySpec").addEventListener("click", async () => {
    await navigator.clipboard.writeText($("specViewer").textContent);
    showToast("📋 Especificación copiada al portapapeles");
  });

  // Codegen & Export
  ["pkgNameInput", "funcNameInput", "chkTypedCodegen"].forEach((id) => {
    $(id).addEventListener("input", debounce(refreshCodegenPreview, 200));
    $(id).addEventListener("change", refreshCodegenPreview);
  });

  $("btnCopyGo").addEventListener("click", async () => {
    await navigator.clipboard.writeText($("codegenPreview").textContent);
    showToast("📋 Código Go copiado al portapapeles");
  });

  $("btnDownloadGo").addEventListener("click", () => {
    const slug = (state.config?.name || "mydsl").toLowerCase().replace(/[^a-z0-9]+/g, "_");
    downloadBlob(`${slug}_gen.go`, $("codegenPreview").textContent, "text/x-go;charset=utf-8");
  });

  $("btnDownloadVSCode").addEventListener("click", async () => {
    const resp = await fetch("/api/studio/vscode.zip", { method: "POST" });
    if (!resp.ok) {
      showToast("✗ Error generando extensión VS Code");
      return;
    }
    const blob = await resp.blob();
    const slug = (state.config?.name || "dsl").toLowerCase().replace(/[^a-z0-9]+/g, "-");
    const a = document.createElement("a");
    a.href = URL.createObjectURL(blob);
    a.download = `${slug}-vscode.zip`;
    a.click();
    URL.revokeObjectURL(a.href);
    showToast("⬇ Extensión VS Code / Cursor (.zip) descargada");
  });

  initStudio();
})();
