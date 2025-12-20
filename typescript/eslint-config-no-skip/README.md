# @claude-code-tools/eslint-config-no-skip

ESLint shareable configuration to detect and error on skipped tests in Jest and Cypress test suites.

This package provides ESLint configurations that prevent accidentally committing skipped tests (`.skip`) or exclusive tests (`.only`) in your test files. It ensures your CI/CD pipeline runs all tests, not just a subset.

## Installation

```bash
pnpm add -D @claude-code-tools/eslint-config-no-skip
```

## Requirements

- ESLint 9.0.0 or higher
- ESLint flat config format (`eslint.config.js`)

## Usage

### Combined (Jest + Cypress)

Import the default configuration to check both Jest and Cypress test files:

```javascript
// eslint.config.js
import noSkipConfig from '@claude-code-tools/eslint-config-no-skip';

export default [
  ...noSkipConfig,
  // your other configs
];
```

This configuration will:
- Apply Jest rules to `**/*.test.{js,ts,jsx,tsx}` and `**/*.spec.{js,ts,jsx,tsx}` files
- Apply Cypress rules to `**/*.cy.{js,ts,jsx,tsx}` and `**/cypress/**/*.{js,ts,jsx,tsx}` files

### Jest Only

Import only the Jest configuration:

```javascript
// eslint.config.js
import jestConfig from '@claude-code-tools/eslint-config-no-skip/jest';

export default [
  ...jestConfig,
  // your other configs
];
```

### Cypress Only

Import only the Cypress configuration:

```javascript
// eslint.config.js
import cypressConfig from '@claude-code-tools/eslint-config-no-skip/cypress';

export default [
  ...cypressConfig,
  // your other configs
];
```

## Detected Patterns

### Jest Tests

| Pattern | Rule | Description |
| --- | --- | --- |
| `describe.skip()` | `jest/no-disabled-tests` | Skipped test suite |
| `it.skip()` | `jest/no-disabled-tests` | Skipped test case |
| `test.skip()` | `jest/no-disabled-tests` | Skipped test case |
| `xdescribe()` | `jest/no-disabled-tests` | Skipped test suite (alternative syntax) |
| `xit()` | `jest/no-disabled-tests` | Skipped test case (alternative syntax) |
| `xtest()` | `jest/no-disabled-tests` | Skipped test case (alternative syntax) |

### Cypress Tests

| Pattern | Rule | Description |
| --- | --- | --- |
| `describe.skip()` | `mocha/no-pending-tests` | Skipped test suite |
| `it.skip()` | `mocha/no-pending-tests` | Skipped test case |
| `context.skip()` | `mocha/no-pending-tests` | Skipped test context |
| `describe.only()` | `mocha/no-exclusive-tests` | Exclusive test suite |
| `it.only()` | `mocha/no-exclusive-tests` | Exclusive test case |
| `context.only()` | `mocha/no-exclusive-tests` | Exclusive test context |

## Troubleshooting

### ESLint 8 or older

This package requires ESLint 9.0.0 or higher with flat config format. If you're using ESLint 8 or older with `.eslintrc.*` files, you'll need to upgrade to ESLint 9 and migrate to flat config.

See the [ESLint migration guide](https://eslint.org/docs/latest/use/configure/migration-guide) for migration instructions.

### Config not being applied

Ensure your test files match the configured patterns:
- Jest: `**/*.test.{js,ts,jsx,tsx}` or `**/*.spec.{js,ts,jsx,tsx}`
- Cypress: `**/*.cy.{js,ts,jsx,tsx}` or `**/cypress/**/*.{js,ts,jsx,tsx}`

If your test files use different naming conventions, you can customize the patterns by creating your own config based on the exported configurations.

## License

MIT
