# Universal Business Rules DSL (Drools-Inspired)

A universal and generic business rules engine inspired by Drools, working with any Go struct type using reflection.

## 🎯 Features

- **100% Universal**: Works with any Go struct using reflection
- **Drools-like Syntax**: Familiar business rules syntax
- **Bilingual**: Supports rules in Spanish and English
- **Salience/Priority**: Rule execution order control
- **Dynamic Context**: Dynamic context support
- **Forward-chaining**: Chain rule execution
- **Working Memory**: Drools-like working memory pattern
- **Struct Tags**: Support for `drools:"fieldname"` tags

## 🚀 Usage

```bash
cd examples/drools
go run main.go
```

## 📖 Rule Syntax

### Basic Structure (Spanish)
```
rule "Rule Name" 
salience PRIORITY 
when ENTITY FIELD OPERATOR VALUE 
then ACTION FIELD a VALUE 
end
```

### Basic Structure (English)
```
rule "Rule Name" 
salience PRIORITY 
when ENTITY FIELD OPERATOR VALUE 
then ACTION FIELD to VALUE 
end
```

## 🔧 Rule Examples

### Spanish Rules
```go
// Customer categorization
rule "Categorizar Cliente Premium" salience 100 
when customer categoria es "regular" and customer balance mayor 10000 
then establecer categoria a "premium" 
end

// High priority orders
rule "Orden Alta Prioridad" salience 70 
when order monto mayor 3000 
then establecer prioridad a "high" 
end

// Inventory control
rule "Stock Bajo" salience 50 
when product stock menor 10 
then establecer estado a "low_stock" 
end
```

### English Rules
```go
// Customer categorization
rule "Premium Customer Discount" salience 30 
when customer categoria is "premium" 
then set estado to "discount_eligible" 
end

// VIP treatment
rule "VIP Customer Special" salience 20 
when customer categoria is "vip" 
then set estado to "vip_treatment" 
end
```

## ⚙️ Supported Operators

### Spanish
- `es` - equality
- `mayor` - greater than
- `menor` - less than
- `contiene` - contains text
- `coincide` - matches pattern

### English
- `is` / `equals` - equality
- `greater` - greater than
- `less` - less than
- `contains` - contains text
- `matches` - matches pattern

## 🎭 Supported Actions

### Spanish
- `establecer` - set field
- `modificar` - modify entity
- `insertar` - insert new entity
- `retirar` - retract entity
- `ejecutar` - execute function

### English
- `set` - set field
- `modify` - modify entity
- `insert` - insert new entity
- `retract` - retract entity
- `execute` - execute function

## 📊 Type Support

The engine works with **any Go struct**:

```go
type Customer struct {
    ID       int     `drools:"id"`
    Name     string  `drools:"nombre"`
    Category string  `drools:"categoria"`
    Balance  float64 `drools:"balance"`
    Status   string  `drools:"estado"`
}

type Product struct {
    ID       int     `drools:"id"`
    Name     string  `drools:"nombre"`
    Price    float64 `drools:"precio"`
    Stock    int     `drools:"stock"`
    Status   string  `drools:"estado"`
}

type Order struct {
    ID         int     `drools:"id"`
    CustomerID int     `drools:"customer_id"`
    Amount     float64 `drools:"monto"`
    Status     string  `drools:"estado"`
    Priority   string  `drools:"prioridad"`
}
```

## 🔄 Execution Cycle

1. **Insert Facts**: Add entities to working memory
2. **Define Rules**: Create business rules with salience
3. **Evaluate Conditions**: Engine evaluates all conditions
4. **Execute Actions**: Actions of matching rules are executed
5. **Forward-chaining**: Modifications can trigger new rules

## 🌟 Salience (Priority)

```go
// Rules executed in priority order (higher salience first)
rule "Critical Rule" salience 100 when ... then ... end     // Executes first
rule "Important Rule" salience 50 when ... then ... end     // Executes second  
rule "Normal Rule" salience 10 when ... then ... end        // Executes third
```

## 🔧 Engine API

### Create Engine
```go
drools := universal.NewUniversalDroolsDSL()
engine := drools.GetEngine()
```

### Insert Facts
```go
customer := &Customer{1, "Juan García", "regular", 5000.0, "active"}
drools.InsertFact(customer)
```

### Define Rules
```go
rule := `rule "My Rule" when customer balance mayor 1000 then establecer categoria a "vip" end`
result, err := drools.Parse(rule)
```

### Execute Rules
```go
err := drools.FireAllRules()
```

### Dynamic Context
```go
context := map[string]interface{}{
    "season": "holiday",
    "promotion": "black_friday",
}
result, err := drools.Use(rule, context)
```

## 🎯 Use Cases

- **E-commerce**: Customer categorization, dynamic discounts
- **Finance**: Credit approval, fraud detection  
- **Inventory**: Stock control, automatic reordering
- **Marketing**: Customer segmentation, personalized campaigns
- **Logistics**: Order prioritization, route optimization

## 🏗️ Architecture

- **UniversalRuleEngine**: Main engine with reflection
- **UniversalDroolsDSL**: Drools syntax parser
- **Rule**: Rule structure with conditions and actions
- **Working Memory**: Facts storage
- **Forward-chaining**: Reactive rule execution

## ✅ Enterprise Features

- ✅ **100% Generic** - Works with any struct
- ✅ **High Performance** - Optimized with efficient reflection  
- ✅ **Natural Syntax** - Similar to natural language
- ✅ **Multilingual** - Spanish and English
- ✅ **Extensible** - Easy to add new operators/actions
- ✅ **Production Ready** - Ready for production
- ✅ **Zero Dependencies** - Only standard Go + go-dsl

A complete and universal business rules engine for Go! 🚀