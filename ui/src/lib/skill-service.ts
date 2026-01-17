// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

import { Skill } from '@proto/config/v1/skill';

const API_BASE = '/v1/skills';

// Helper to wrap skill in request object if needed, or just send valid JSON that matches Proto.
// The generated Gateway handlers expects JSON body mapped to the request message.
// CreateSkillRequest has field 'skill'.
// UpdateSkillRequest has field 'skill'.
// So we should wrap the skill in { skill: ... } object?
// "body: 'skill'" option in proto means the body content is mapped TO the 'skill' field.
// So sending the Skill object directly is correct.

/**
 * SkillService provides methods to interact with the Skill API.
 */
export const SkillService = {
  async list(): Promise<Skill[]> {
    const res = await fetch(API_BASE);
    if (!res.ok) throw new Error('Failed to list skills');
    /*
      Proto JSON response for ListSkillsResponse:
      { "skills": [...] }
    */
    const data = await res.json();
    return data.skills || [];
  },

  async get(name: string): Promise<Skill> {
    const res = await fetch(`${API_BASE}/${name}`);
    if (!res.ok) throw new Error(`Failed to get skill ${name}`);
    /*
      Proto JSON response for GetSkillResponse:
      { "skill": {...} }
    */
    const data = await res.json();
    return data.skill;
  },

  async create(skill: Skill): Promise<Skill> {
    const res = await fetch(API_BASE, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(skill),
    });
    if (!res.ok) {
        const err = await res.text();
        throw new Error(`Failed to create skill: ${err}`);
    }
    const data = await res.json();
    return data.skill;
  },

  async update(originalName: string, skill: Skill): Promise<Skill> {
    const res = await fetch(`${API_BASE}/${originalName}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(skill),
    });
    if (!res.ok) {
        const err = await res.text();
        throw new Error(`Failed to update skill: ${err}`);
    }
    const data = await res.json();
    return data.skill;
  },

  async delete(name: string): Promise<void> {
    const res = await fetch(`${API_BASE}/${name}`, {
      method: 'DELETE',
    });
    if (!res.ok) throw new Error(`Failed to delete skill ${name}`);
  },

  async uploadAsset(skillName: string, path: string, file: File): Promise<void> {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('path', path); // We might need to handle path in backend or query param

    // Use /v1/skills/{name}/assets
    const res = await fetch(`${API_BASE}/${skillName}/assets`, {
      method: 'POST',
      body: formData,
    });

    if (!res.ok) {
        const err = await res.text();
        throw new Error(`Failed to upload asset: ${err}`);
    }
  },
};

export type { Skill };
