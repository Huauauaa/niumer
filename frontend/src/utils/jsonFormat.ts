/** Trim trailing semicolons often pasted after JS object literals. */
function trimInput(input: string): string {
  return input.trim().replace(/;+\s*$/u, "");
}

/**
 * Parse strict JSON first; if that fails, evaluate as a JavaScript expression
 * (object/array literal with unquoted keys, single quotes, etc.).
 * Intended for local-only tooling (same tradeoff as Postman-style formatters).
 */
export function parseJsonOrJsLiteral(input: string): unknown {
  const raw = trimInput(input);
  if (!raw) {
    throw new Error("Empty input");
  }
  try {
    return JSON.parse(raw);
  } catch {
    try {
      return new Function(`"use strict"; return (${raw});`)();
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e);
      throw new Error(`Invalid JSON or JavaScript literal: ${msg}`);
    }
  }
}

export function stringifyPretty(value: unknown): string {
  try {
    const s = JSON.stringify(value, null, 2);
    if (s === undefined) {
      throw new Error("Result cannot be represented as JSON (e.g. undefined).");
    }
    return s;
  } catch (e) {
    if (e instanceof TypeError) {
      throw new Error(
        "Value cannot be serialized to JSON (e.g. BigInt or circular structure).",
      );
    }
    throw e;
  }
}

export function stringifyMinified(value: unknown): string {
  try {
    const s = JSON.stringify(value);
    if (s === undefined) {
      throw new Error("Result cannot be represented as JSON (e.g. undefined).");
    }
    return s;
  } catch (e) {
    if (e instanceof TypeError) {
      throw new Error(
        "Value cannot be serialized to JSON (e.g. BigInt or circular structure).",
      );
    }
    throw e;
  }
}
