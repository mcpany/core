/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */
<<<<<<< HEAD
=======

>>>>>>> 4f4bb883 (chore: verification fixes (lint, tests, ui, dep updates))


import { NextResponse } from 'next/server';

export async function GET() {
  const prompts = [
    { name: "summarize_text", description: "Summarizes the given text", service: "LLMService" },
    { name: "code_review", description: "Reviews code for bugs", service: "LLMService" },
  ];
  return NextResponse.json(prompts);
}
