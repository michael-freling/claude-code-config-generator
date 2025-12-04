---
name: coding
description: Iterative coding development with Test-Driven Development (TDD). Follows a structured workflow of planning reviewable changes, implementing with tests, getting code review, and committing incrementally. Use when implementing features or changes that require iterative development with verification at each step.
allowed-tools: [Read, Write, Edit, Glob, Grep, Bash]
---

# Iterative Coding Development Skill (TDD)

A comprehensive skill for iterative software development following Test-Driven Development principles. This skill emphasizes planning reviewable changes, implementing with verification, code review, and incremental commits.

## When to Use

Use this skill when:
- Implementing features that require multiple coordinated changes
- Working on changes that need verification at each step
- Following Test-Driven Development practices
- Need structured code review workflow
- Building reliable, tested software incrementally
- Refactoring existing code with safety nets

## Process

### 0. Read Project Design Documentation

**CRITICAL FIRST STEP: Always check for and read `.claude/docs/guideline.md`**

Before starting any implementation:

1. **Look for `.claude/docs/guideline.md` in the current directory**
   - If found, read it thoroughly
   - This contains project-specific coding standards, conventions, and architecture
   - Follow these guidelines strictly as they override general best practices

2. **For monorepos or subprojects:**
   - Check for `.claude/docs/guideline.md` in the subproject root
   - Also check the repository root for overall standards
   - Subproject-specific rules take precedence over repository-level rules

3. **If no guideline.md exists:**
   - Consider running `/document-guideline` to create one
   - Or proceed with analyzing the codebase manually

**What to extract from guideline.md:**
- Project-specific coding conventions
- Testing framework and patterns
- Directory structure and organization
- Naming conventions
- Error handling patterns
- Code examples showing preferred style

### 1. Plan Reviewable Changes

**Split the work into logical, reviewable units:**

Before writing any code, analyze the requirements and break them into discrete changes:

1. **Identify Change Groups**
   - Each group should be independently implementable
   - Each group should be independently verifiable and testable
   - Each group should be small enough for meaningful review
   - Changes should build on each other logically

2. **Define Success Criteria for Each Change**
   - What tests need to pass?
   - What behavior should be observable?
   - What edge cases need handling?

3. **Order Changes by Dependencies**
   - Start with foundational changes
   - Build up incrementally
   - Each change should leave the codebase in a working state

Example Planning:
```
Change 1: Add data model and validation
  - Tests: Unit tests for validation rules
  - Verify: All validation tests pass

Change 2: Implement core business logic
  - Tests: Unit tests for business rules
  - Verify: Business logic tests pass

Change 3: Add API endpoint
  - Tests: Integration tests for endpoint
  - Verify: API returns expected responses

Change 4: Wire up to UI
  - Tests: E2E tests for user flow
  - Verify: Full flow works end-to-end
```

### 2. Implement Each Change (TDD Cycle)

For each planned change, follow the TDD cycle:

**Step 1: Write Tests First**
- Write failing tests that define expected behavior
- Include both happy path and error cases
- Use table-driven tests for comprehensive coverage

**Step 2: Implement Code**
- Write minimal code to make tests pass
- Follow coding best practices (see below)
- Handle all errors explicitly

**Step 3: Verify**
- Run all tests to confirm they pass
- Run linters and type checkers
- Build the project to catch compilation errors
- Manual verification if applicable

**Step 4: Refactor**
- Clean up code while keeping tests green
- Remove duplication
- Improve naming and structure

### 3. Get Code Review

After completing a change:

1. **Self-Review**
   - Review your own changes for issues
   - Check for missed edge cases
   - Verify test coverage is adequate

2. **Request Review from Reviewer Agent**
   - Submit changes for review
   - Address all feedback
   - Re-verify after making changes

3. **Iterate Until Approved**
   - Make requested improvements
   - Add additional tests if needed
   - Document any non-obvious decisions

### 4. Commit the Change

Only commit after:
- All tests pass
- All linters pass
- Build succeeds
- Review is complete

Create a meaningful commit message that describes:
- What changed
- Why it changed
- Any relevant context

### 5. Repeat Until Complete

Move to the next change in the plan and repeat steps 2-4 until all changes are implemented.

## Core Principles

### Simplicity First
- **DRY (Don't Repeat Yourself)**: Extract common patterns; update existing code for reusability rather than creating new
- **Early Returns**: Prefer `continue` or `return early` over nested conditionals - "if is bad, else is worse"
- **Minimal Scope**: Keep variables local when possible
- **Delete Dead Code**: Remove unused code immediately; version control preserves history

### Code Quality
- **Every Error Must Be Checked or Returned**: Never silently ignore errors
- **No Hacks, No Assumptions, No Global State**: Write clean, predictable code
- **Always Default to Production-Safe Behavior**: Code should be safe by default

### Environment Independence
- **Write code that operates identically across dev, test, and production**
- **Avoid environment-specific logic in core logic**: Use configuration or dependency injection
- **Use test doubles externally, not via conditionals in production code**

### Comments
- **Comments MUST BE about WHY not WHAT** - Explain reasoning behind decisions
- **Minimal comments**: Only high-level explanations of purpose, architecture, or non-obvious decisions
- **No line-by-line comments**: Code should be self-documenting

## Coding Best Practices

### Error Handling

```go
// Good: Early return, explicit error handling
func ProcessUser(id string) (*User, error) {
    if id == "" {
        return nil, errors.New("user ID is required")
    }

    user, err := db.GetUser(id)
    if err != nil {
        return nil, fmt.Errorf("failed to get user: %w", err)
    }

    if !user.IsActive {
        return nil, ErrUserInactive
    }

    return user, nil
}

// Bad: Nested conditionals, swallowed errors
func ProcessUser(id string) *User {
    if id != "" {
        user, err := db.GetUser(id)
        if err == nil {
            if user.IsActive {
                return user
            }
        }
    }
    return nil  // Error context lost
}
```

### Early Returns Over Nesting

```typescript
// Good: Early returns, flat structure
function validateOrder(order: Order): ValidationResult {
    if (!order) {
        return { valid: false, error: 'Order is required' };
    }

    if (!order.items || order.items.length === 0) {
        return { valid: false, error: 'Order must have items' };
    }

    if (order.total <= 0) {
        return { valid: false, error: 'Order total must be positive' };
    }

    return { valid: true };
}

// Bad: Nested conditionals
function validateOrder(order: Order): ValidationResult {
    if (order) {
        if (order.items && order.items.length > 0) {
            if (order.total > 0) {
                return { valid: true };
            } else {
                return { valid: false, error: 'Order total must be positive' };
            }
        } else {
            return { valid: false, error: 'Order must have items' };
        }
    } else {
        return { valid: false, error: 'Order is required' };
    }
}
```

### Configuration Over Environment Branching

```python
# Good: Dependency injection, configuration-driven
class PaymentService:
    def __init__(self, payment_gateway: PaymentGateway):
        self.gateway = payment_gateway

    def process(self, payment: Payment) -> Result:
        return self.gateway.charge(payment)

# In production
service = PaymentService(StripeGateway(config.stripe_key))

# In tests
service = PaymentService(MockPaymentGateway())

# Bad: Environment checks in production code
class PaymentService:
    def process(self, payment: Payment) -> Result:
        if os.environ.get('ENV') == 'test':
            return Result(success=True)  # Fake in tests
        elif os.environ.get('ENV') == 'development':
            return self._dev_charge(payment)
        else:
            return self._prod_charge(payment)
```

### Clean Code Patterns

```java
// Good: Single responsibility, clear naming
public class OrderValidator {
    private final PriceCalculator priceCalculator;
    private final InventoryChecker inventoryChecker;

    public ValidationResult validate(Order order) {
        var priceResult = validatePricing(order);
        if (!priceResult.isValid()) {
            return priceResult;
        }

        return validateInventory(order);
    }

    private ValidationResult validatePricing(Order order) {
        var calculatedTotal = priceCalculator.calculate(order.getItems());
        if (!calculatedTotal.equals(order.getTotal())) {
            return ValidationResult.invalid("Price mismatch");
        }
        return ValidationResult.valid();
    }

    private ValidationResult validateInventory(Order order) {
        for (var item : order.getItems()) {
            if (!inventoryChecker.isAvailable(item)) {
                return ValidationResult.invalid("Item unavailable: " + item.getId());
            }
        }
        return ValidationResult.valid();
    }
}
```

## Testing

### Table-Driven Testing

**CRITICAL: Always use table-driven tests as the primary testing approach.**

Table-driven tests provide:
- Reduced code duplication
- Easy addition of new test cases
- Improved readability
- Consistent test structure
- Obvious coverage gaps

### Separate Happy Path and Error Cases

**IMPORTANT: Split test cases into success and error scenarios**

```go
func TestValidateUser(t *testing.T) {
    // Success test cases
    successTestCases := []struct {
        name     string
        input    User
        expected bool
    }{
        {
            name:     "valid user with all fields",
            input:    User{Name: "John", Email: "john@example.com", Age: 25},
            expected: true,
        },
        {
            name:     "valid user with minimum age",
            input:    User{Name: "Jane", Email: "jane@example.com", Age: 18},
            expected: true,
        },
    }

    for _, tc := range successTestCases {
        t.Run(tc.name, func(t *testing.T) {
            result := ValidateUser(tc.input)
            if result != tc.expected {
                t.Errorf("expected %v, got %v", tc.expected, result)
            }
        })
    }

    // Error test cases
    errorTestCases := []struct {
        name          string
        input         User
        expectedError string
    }{
        {
            name:          "empty name",
            input:         User{Name: "", Email: "test@example.com", Age: 25},
            expectedError: "name is required",
        },
        {
            name:          "invalid email",
            input:         User{Name: "John", Email: "invalid", Age: 25},
            expectedError: "invalid email format",
        },
        {
            name:          "underage user",
            input:         User{Name: "John", Email: "john@example.com", Age: 17},
            expectedError: "user must be at least 18",
        },
    }

    for _, tc := range errorTestCases {
        t.Run(tc.name, func(t *testing.T) {
            err := ValidateUserWithError(tc.input)
            if err == nil {
                t.Fatal("expected error but got nil")
            }
            if !strings.Contains(err.Error(), tc.expectedError) {
                t.Errorf("expected error containing %q, got %q", tc.expectedError, err.Error())
            }
        })
    }
}
```

### Test Case Design Principles

1. **Define test inputs as test case fields, not function arguments**
```go
// Good: Input is part of test case struct
type testCase struct {
    name     string
    input    InputData
    expected OutputData
}

// Bad: Input defined outside test case
func TestFunction(t *testing.T) {
    input := createInput()  // Shared input is harder to trace
    // ...
}
```

2. **Avoid redundant test cases with the same purpose**
```go
// Bad: Redundant cases testing same thing
{name: "valid email 1", email: "a@b.com"},
{name: "valid email 2", email: "x@y.com"},
{name: "valid email 3", email: "test@example.com"},

// Good: Each case tests distinct behavior
{name: "simple valid email", email: "a@b.com"},
{name: "email with subdomain", email: "user@sub.example.com"},
{name: "email with plus addressing", email: "user+tag@example.com"},
```

3. **Prefer injecting values over changing global state**
```go
// Good: Inject dependencies
func TestService(t *testing.T) {
    mockDB := NewMockDB()
    service := NewService(mockDB)
    // ...
}

// Bad: Modify global state
func TestService(t *testing.T) {
    oldDB := globalDB
    globalDB = mockDB
    defer func() { globalDB = oldDB }()
    // ...
}
```

### TypeScript/JavaScript Table-Driven Tests

```typescript
describe('calculateDiscount', () => {
    // Success test cases
    const successTestCases = [
        {
            name: 'applies 10% discount for orders over $100',
            orderTotal: 150,
            customerTier: 'standard',
            expected: 15,
        },
        {
            name: 'applies 20% discount for premium customers',
            orderTotal: 100,
            customerTier: 'premium',
            expected: 20,
        },
        {
            name: 'no discount for orders under threshold',
            orderTotal: 50,
            customerTier: 'standard',
            expected: 0,
        },
    ];

    successTestCases.forEach((tc) => {
        it(tc.name, () => {
            const result = calculateDiscount(tc.orderTotal, tc.customerTier);
            expect(result).toBe(tc.expected);
        });
    });

    // Error test cases
    const errorTestCases = [
        {
            name: 'throws error for negative order total',
            orderTotal: -10,
            customerTier: 'standard',
            expectedError: 'Order total must be positive',
        },
        {
            name: 'throws error for invalid customer tier',
            orderTotal: 100,
            customerTier: 'invalid',
            expectedError: 'Invalid customer tier',
        },
    ];

    errorTestCases.forEach((tc) => {
        it(tc.name, () => {
            expect(() => calculateDiscount(tc.orderTotal, tc.customerTier))
                .toThrow(tc.expectedError);
        });
    });
});
```

### Python Table-Driven Tests

```python
import pytest

class TestValidatePassword:
    # Success test cases
    success_test_cases = [
        {
            "name": "valid password with all requirements",
            "password": "SecurePass123!",
            "expected": True,
        },
        {
            "name": "valid password at minimum length",
            "password": "Abc123!@",
            "expected": True,
        },
    ]

    @pytest.mark.parametrize("tc", success_test_cases, ids=lambda tc: tc["name"])
    def test_success_cases(self, tc):
        result = validate_password(tc["password"])
        assert result == tc["expected"]

    # Error test cases
    error_test_cases = [
        {
            "name": "password too short",
            "password": "Ab1!",
            "expected_error": "Password must be at least 8 characters",
        },
        {
            "name": "password missing uppercase",
            "password": "lowercase123!",
            "expected_error": "Password must contain uppercase letter",
        },
        {
            "name": "password missing special character",
            "password": "SecurePass123",
            "expected_error": "Password must contain special character",
        },
    ]

    @pytest.mark.parametrize("tc", error_test_cases, ids=lambda tc: tc["name"])
    def test_error_cases(self, tc):
        with pytest.raises(ValidationError) as exc_info:
            validate_password(tc["password"])
        assert tc["expected_error"] in str(exc_info.value)
```

## Implementation Strategy

### Step 1: Review Project Guidelines
- Read `.claude/docs/guideline.md` if it exists (MANDATORY)
- Extract project-specific patterns
- Note testing framework and conventions
- Identify error handling patterns

### Step 2: Plan the Changes
- Break work into reviewable units
- Define success criteria for each
- Order by dependencies
- Ensure each change is independently testable

### Step 3: For Each Change, Follow TDD

**Write Tests First:**
```bash
# Run tests - they should fail
go test ./...
# or
npm test
# or
pytest
```

**Implement Code:**
- Write minimal code to pass tests
- Follow coding best practices
- Handle all errors explicitly

**Verify Everything Passes:**
```bash
# Run all quality checks
go test ./... && go vet ./... && golangci-lint run
# or
npm test && npm run lint && npm run build
# or
pytest && flake8 && mypy .
```

**Refactor:**
- Clean up while keeping tests green
- Remove duplication
- Improve naming

### Step 4: Get Review

Submit for code review and address all feedback:
- Fix any issues identified
- Add tests for missed edge cases
- Document non-obvious decisions

### Step 5: Commit

Only after all checks pass:
```bash
git add -A
git commit -m "feat: implement [change description]

- Add [specific change 1]
- Add [specific change 2]
- Tests cover [scenarios]"
```

### Step 6: Repeat

Move to next planned change and repeat until complete.

## Checklist

### Before Starting
- [ ] **Read `.claude/docs/guideline.md` if it exists** (CRITICAL)
- [ ] Understand the full requirements
- [ ] Analyze existing code patterns
- [ ] Identify testing framework and conventions
- [ ] Plan changes into reviewable units

### For Each Change

**Before Implementing:**
- [ ] Define success criteria
- [ ] Identify test cases (happy path and errors)
- [ ] Understand dependencies on previous changes

**During Implementation:**
- [ ] Write tests first (TDD)
- [ ] Implement minimal code to pass tests
- [ ] Use early returns over nested conditionals
- [ ] Handle all errors explicitly
- [ ] Use configuration/DI over environment branching
- [ ] Write minimal comments (WHY not WHAT)
- [ ] Delete dead code

**After Implementation:**
- [ ] All tests pass
- [ ] Linter passes
- [ ] Type checker passes
- [ ] Build succeeds
- [ ] Self-review completed
- [ ] Reviewer feedback addressed
- [ ] Change committed

### After All Changes Complete
- [ ] Full test suite passes
- [ ] All linters pass
- [ ] Build succeeds in all environments
- [ ] Documentation updated if needed
- [ ] All commits are clean and meaningful

## Key Principles Summary

1. **Project Guidelines First**: Always read and follow `.claude/docs/guideline.md`
2. **Plan Before Code**: Split work into reviewable, testable changes
3. **TDD Cycle**: Write tests first, implement, verify, refactor
4. **Early Returns**: "if is bad, else is worse" - avoid nesting
5. **Explicit Error Handling**: Every error must be checked or returned
6. **Environment Independence**: Code works identically in dev, test, and production
7. **Configuration Over Branching**: Use DI and configuration, not environment checks
8. **Minimal Comments**: Only WHY, not WHAT; no line-by-line comments
9. **Delete Dead Code**: Version control preserves history
10. **Table-Driven Tests**: Separate happy path and error cases
11. **No Hacks**: No assumptions, no global state, production-safe defaults
12. **Review Before Commit**: Get review, address feedback, then commit

## Version History

### Version 1.0 (2025-01-16)
- Initial iterative TDD coding skill
- Structured workflow: plan, implement, review, commit
- Comprehensive coding best practices
- Table-driven testing patterns with separate success/error cases
- Error handling and early return patterns
- Environment independence and dependency injection
- Multi-language examples (Go, TypeScript, Python, Java)
- Complete checklist for iterative development
