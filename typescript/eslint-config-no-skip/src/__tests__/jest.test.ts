import { describe, it, expect } from 'vitest';
import { jestConfig } from '../jest.js';

describe('jestConfig', () => {
  it('should export an array of configs', () => {
    expect(Array.isArray(jestConfig)).toBe(true);
    expect(jestConfig.length).toBeGreaterThan(0);
  });

  describe('config structure', () => {
    const testCases = [
      {
        name: 'should have name property',
        assertion: (config: typeof jestConfig[0]) => {
          expect(config.name).toBe('claude-code-tools/no-skip/jest');
        },
      },
      {
        name: 'should have files array',
        assertion: (config: typeof jestConfig[0]) => {
          expect(Array.isArray(config.files)).toBe(true);
          expect(config.files.length).toBeGreaterThan(0);
        },
      },
      {
        name: 'should have plugins object',
        assertion: (config: typeof jestConfig[0]) => {
          expect(config.plugins).toBeDefined();
          expect(typeof config.plugins).toBe('object');
        },
      },
      {
        name: 'should have rules object',
        assertion: (config: typeof jestConfig[0]) => {
          expect(config.rules).toBeDefined();
          expect(typeof config.rules).toBe('object');
        },
      },
      {
        name: 'should have languageOptions with globals',
        assertion: (config: typeof jestConfig[0]) => {
          expect(config.languageOptions).toBeDefined();
          expect(config.languageOptions?.globals).toBeDefined();
        },
      },
    ];

    testCases.forEach(({ name, assertion }) => {
      it(name, () => {
        const config = jestConfig[0];
        assertion(config);
      });
    });
  });

  describe('file patterns', () => {
    const expectedPatterns = [
      { pattern: '**/*.test.{js,ts,jsx,tsx}', description: 'test files' },
      { pattern: '**/*.spec.{js,ts,jsx,tsx}', description: 'spec files' },
    ];

    expectedPatterns.forEach(({ pattern, description }) => {
      it(`should include pattern for ${description}: ${pattern}`, () => {
        const config = jestConfig[0];
        expect(config.files).toContain(pattern);
      });
    });
  });

  describe('plugin configuration', () => {
    it('should have jest plugin configured', () => {
      const config = jestConfig[0];
      expect(config.plugins?.jest).toBeDefined();
    });
  });

  describe('language options', () => {
    const expectedGlobals = [
      { name: 'describe', type: 'readonly' },
      { name: 'it', type: 'readonly' },
      { name: 'test', type: 'readonly' },
      { name: 'expect', type: 'readonly' },
      { name: 'beforeAll', type: 'readonly' },
      { name: 'afterAll', type: 'readonly' },
      { name: 'beforeEach', type: 'readonly' },
      { name: 'afterEach', type: 'readonly' },
      { name: 'jest', type: 'readonly' },
    ];

    expectedGlobals.forEach(({ name, type }) => {
      it(`should define ${name} as ${type}`, () => {
        const config = jestConfig[0];
        expect(config.languageOptions?.globals?.[name]).toBe(type);
      });
    });
  });

  describe('rules configuration', () => {
    const expectedRules = [
      {
        rule: 'jest/no-disabled-tests',
        level: 'error',
        description: 'should error on disabled tests',
      },
    ];

    expectedRules.forEach(({ rule, level, description }) => {
      it(`${description} (${rule} = ${level})`, () => {
        const config = jestConfig[0];
        expect(config.rules?.[rule]).toBe(level);
      });
    });
  });
});
