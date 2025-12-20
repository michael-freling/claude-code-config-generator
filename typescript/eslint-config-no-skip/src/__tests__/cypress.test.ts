import { describe, it, expect } from 'vitest';
import { cypressConfig } from '../cypress.js';

describe('cypressConfig', () => {
  it('should export an array of configs', () => {
    expect(Array.isArray(cypressConfig)).toBe(true);
    expect(cypressConfig.length).toBeGreaterThan(0);
  });

  describe('config structure', () => {
    const testCases = [
      {
        name: 'should have name property',
        assertion: (config: typeof cypressConfig[0]) => {
          expect(config.name).toBe('claude-code-tools/no-skip/cypress');
        },
      },
      {
        name: 'should have files array',
        assertion: (config: typeof cypressConfig[0]) => {
          expect(Array.isArray(config.files)).toBe(true);
          expect(config.files.length).toBeGreaterThan(0);
        },
      },
      {
        name: 'should have plugins object',
        assertion: (config: typeof cypressConfig[0]) => {
          expect(config.plugins).toBeDefined();
          expect(typeof config.plugins).toBe('object');
        },
      },
      {
        name: 'should have rules object',
        assertion: (config: typeof cypressConfig[0]) => {
          expect(config.rules).toBeDefined();
          expect(typeof config.rules).toBe('object');
        },
      },
    ];

    testCases.forEach(({ name, assertion }) => {
      it(name, () => {
        const config = cypressConfig[0];
        assertion(config);
      });
    });
  });

  describe('file patterns', () => {
    const expectedPatterns = [
      { pattern: '**/*.cy.{js,ts,jsx,tsx}', description: 'Cypress test files' },
      { pattern: '**/cypress/**/*.{js,ts,jsx,tsx}', description: 'files in cypress directory' },
    ];

    expectedPatterns.forEach(({ pattern, description }) => {
      it(`should include pattern for ${description}: ${pattern}`, () => {
        const config = cypressConfig[0];
        expect(config.files).toContain(pattern);
      });
    });
  });

  describe('plugin configuration', () => {
    it('should have mocha plugin configured', () => {
      const config = cypressConfig[0];
      expect(config.plugins?.mocha).toBeDefined();
    });
  });

  describe('rules configuration', () => {
    const expectedRules = [
      {
        rule: 'mocha/no-pending-tests',
        level: 'error',
        description: 'should error on pending tests',
      },
      {
        rule: 'mocha/no-exclusive-tests',
        level: 'error',
        description: 'should error on exclusive tests',
      },
    ];

    expectedRules.forEach(({ rule, level, description }) => {
      it(`${description} (${rule} = ${level})`, () => {
        const config = cypressConfig[0];
        expect(config.rules?.[rule]).toBe(level);
      });
    });
  });
});
