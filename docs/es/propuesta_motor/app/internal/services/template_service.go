package services

import (
	"fmt"
	"time"

	"motor-contable-poc/internal/models"
	"motor-contable-poc/internal/repository"
	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

type TemplateService struct {
	repository      repository.TemplateRepository
	journalService  *JournalEntryService
	accountService  *AccountService
	dsl            *dslbuilder.DSL
}

func NewTemplateService(
	repo repository.TemplateRepository,
	journalService *JournalEntryService,
	accountService *AccountService,
) *TemplateService {
	service := &TemplateService{
		repository:     repo,
		journalService: journalService,
		accountService: accountService,
	}
	service.initializeDSL()
	return service
}

func (s *TemplateService) initializeDSL() {
	// Initialize DSL for template processing
	s.dsl = dslbuilder.New("TemplateEngine")
	
	// Define tokens for template DSL
	s.dsl.Token("TEMPLATE", `template`)
	s.dsl.Token("PARAMS", `params`)
	s.dsl.Token("ENTRY", `entry`)
	s.dsl.Token("LINE", `line`)
	s.dsl.Token("DEBIT", `debit`)
	s.dsl.Token("CREDIT", `credit`)
	s.dsl.Token("ACCOUNT", `account`)
	s.dsl.Token("AMOUNT", `amount`)
	s.dsl.Token("DESCRIPTION", `description`)
	s.dsl.Token("DATE", `date`)
	s.dsl.Token("VARIABLE", `\$[a-zA-Z_][a-zA-Z0-9_]*`)
	s.dsl.Token("STRING", `"[^"]*"`)
	s.dsl.Token("NUMBER", `\d+(\.\d+)?`)
	s.dsl.Token("LPAREN", `\(`)
	s.dsl.Token("RPAREN", `\)`)
	
	// Register functions
	s.registerDSLFunctions()
}

func (s *TemplateService) registerDSLFunctions() {
	// Function to get last day of period
	s.dsl.Action("last_day", func(args []interface{}) (interface{}, error) {
		if len(args) == 0 {
			return time.Now().Format("2006-01-02"), nil
		}
		period := args[0].(string)
		// Parse period and return last day
		return fmt.Sprintf("%s-31", period), nil
	})
	
	// Function to get current date
	s.dsl.Action("current_date", func(args []interface{}) (interface{}, error) {
		return time.Now().Format("2006-01-02"), nil
	})
	
	// Add more template functions as needed
}

// Create creates a new template
func (s *TemplateService) Create(template *models.JournalTemplate, userID string) error {
	template.CreatedBy = userID
	return s.repository.Create(template)
}

// Update updates an existing template
func (s *TemplateService) Update(template *models.JournalTemplate) error {
	return s.repository.Update(template)
}

// Delete deletes a template
func (s *TemplateService) Delete(id string) error {
	return s.repository.Delete(id)
}

// GetByID retrieves a template by ID
func (s *TemplateService) GetByID(id string) (*models.JournalTemplate, error) {
	return s.repository.GetByID(id)
}

// GetAll retrieves all templates for an organization
func (s *TemplateService) GetAll(organizationID string) ([]models.JournalTemplate, error) {
	return s.repository.GetAll(organizationID)
}

// PreviewTemplate executes a template in preview mode
func (s *TemplateService) PreviewTemplate(templateID string, params map[string]interface{}) (*models.JournalEntry, error) {
	template, err := s.repository.GetByID(templateID)
	if err != nil {
		return nil, fmt.Errorf("template not found: %w", err)
	}
	
	// Create a mock journal entry for preview
	entry := &models.JournalEntry{
		Date:         time.Now(),
		Description:  "Preview - " + template.Description,
		Status:       "PREVIEW",
		JournalLines: []models.JournalLine{},
	}
	
	// For now, create sample lines based on template
	// In real implementation, this would parse and execute the DSL
	if amount, ok := params["amount"].(float64); ok {
		entry.JournalLines = append(entry.JournalLines, models.JournalLine{
			AccountID:    "1105", // Cash account
			Description:  "Sample debit line",
			DebitAmount:  amount,
			CreditAmount: 0,
		})
		
		entry.JournalLines = append(entry.JournalLines, models.JournalLine{
			AccountID:    "4105", // Revenue account
			Description:  "Sample credit line",
			DebitAmount:  0,
			CreditAmount: amount,
		})
	}
	
	// Calculate totals
	entry.CalculateTotals()
	
	return entry, nil
}

// ExecuteTemplate executes a template and creates a journal entry
func (s *TemplateService) ExecuteTemplate(templateID string, params map[string]interface{}, userID string) (*models.JournalEntry, error) {
	// For POC, just preview the template
	// In real implementation, this would:
	// 1. Execute the DSL with parameters
	// 2. Generate journal entry
	// 3. Save to database
	// 4. Record execution history
	
	entry, err := s.PreviewTemplate(templateID, params)
	if err != nil {
		return nil, err
	}
	
	// Set proper values for demo
	entry.ID = "je-" + fmt.Sprintf("%d", time.Now().Unix())
	entry.EntryNumber = fmt.Sprintf("JE-%06d", time.Now().Unix()%1000000)
	entry.Status = "POSTED"
	entry.PostedAt = time.Now()
	
	return entry, nil
}

// GetExecutionHistory retrieves execution history for a template
func (s *TemplateService) GetExecutionHistory(templateID string, limit int) ([]models.TemplateExecution, error) {
	return s.repository.GetExecutionHistory(templateID, limit)
}