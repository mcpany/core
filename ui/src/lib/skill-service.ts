// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

import { apiClient } from './client';
import { Skill } from '@proto/config/v1/skill';

/**
 * SkillService provides methods to interact with the backend Skills API.
 * It handles listing, getting, creating, updating, and deleting skills.
 * It delegates to the centralized apiClient to ensure authentication consistency.
 */
export const SkillService = {
  /**
   * Lists all available skills.
   *
   * @returns A promise that resolves to an array of skills.
   * @throws Error if the request fails.
   */
  async list(): Promise<Skill[]> {
    return apiClient.listSkills();
  },

  /**
   * Retrieves a specific skill by name.
   *
   * @param name - The name of the skill to retrieve.
   * @returns A promise that resolves to the skill object.
   * @throws Error if the request fails.
   */
  async get(name: string): Promise<Skill> {
    return apiClient.getSkill(name);
  },

  /**
   * Creates a new skill.
   *
   * @param skill - The skill object to create.
   * @returns A promise that resolves to the created skill.
   * @throws Error if the request fails.
   */
  async create(skill: Skill): Promise<Skill> {
    return apiClient.createSkill(skill);
  },

  /**
   * Updates an existing skill.
   *
   * @param originalName - The original name of the skill (in case the name is being updated).
   * @param skill - The updated skill object.
   * @returns A promise that resolves to the updated skill.
   * @throws Error if the request fails.
   */
  async update(originalName: string, skill: Skill): Promise<Skill> {
    return apiClient.updateSkill(originalName, skill);
  },

  /**
   * Deletes a skill.
   *
   * @param name - The name of the skill to delete.
   * @returns A promise that resolves when the operation is complete.
   * @throws Error if the request fails.
   */
  async delete(name: string): Promise<void> {
    return apiClient.deleteSkill(name);
  },

  /**
   * Uploads an asset for a skill.
   *
   * @param skillName - The name of the skill.
   * @param path - The target path for the asset.
   * @param file - The file to upload.
   * @returns A promise that resolves when the upload is complete.
   * @throws Error if the request fails.
   */
  async uploadAsset(skillName: string, path: string, file: File): Promise<void> {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('path', path);

    const res = await fetch(`/api/v1/skills/${skillName}/assets`, {
      method: 'POST',
      body: formData,
      // fetchWithAuth doesn't handle FormData easily yet, and this endpoint
      // might be handled differently. If needed, we can add specialized
      // fetchWithAuthFormData later.
    });

    if (!res.ok) {
        const err = await res.text();
        throw new Error(`Failed to upload asset: ${err}`);
    }
  },
};

export type { Skill };
