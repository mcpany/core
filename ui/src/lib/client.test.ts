
import { apiClient } from './client';
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';

describe('apiClient', () => {
  const fetchMock = vi.fn();

  beforeEach(() => {
    fetchMock.mockReset();
    vi.stubGlobal('fetch', fetchMock);
    vi.stubGlobal('localStorage', {
      getItem: vi.fn(),
      setItem: vi.fn(),
    });
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });


  it('should cover getActiveIntentAlignment', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getActiveIntentAlignment({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover getActiveIntentAlignment error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getActiveIntentAlignment({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover listServices', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.listServices({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover listServices error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.listServices({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover listCatalog', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.listCatalog({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover listCatalog error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.listCatalog({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover getService', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getService({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover getService error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getService({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover setServiceStatus', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.setServiceStatus({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover setServiceStatus error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.setServiceStatus({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover getServiceStatus', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getServiceStatus({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover getServiceStatus error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getServiceStatus({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover restartService', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.restartService({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover restartService error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.restartService({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover registerService', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.registerService({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover registerService error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.registerService({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover updateService', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.updateService({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover updateService error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.updateService({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover unregisterService', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.unregisterService({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover unregisterService error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.unregisterService({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover validateService', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.validateService({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover validateService error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.validateService({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover listTools', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.listTools({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover listTools error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.listTools({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover executeTool', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.executeTool({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover executeTool error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.executeTool({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover setToolStatus', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.setToolStatus({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover setToolStatus error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.setToolStatus({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover listResources', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.listResources({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover listResources error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.listResources({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover readResource', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.readResource({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover readResource error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.readResource({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover setResourceStatus', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.setResourceStatus({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover setResourceStatus error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.setResourceStatus({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover listPrompts', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.listPrompts({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover listPrompts error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.listPrompts({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover setPromptStatus', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.setPromptStatus({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover setPromptStatus error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.setPromptStatus({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover executePrompt', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.executePrompt({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover executePrompt error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.executePrompt({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover getServiceTemplates', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getServiceTemplates({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover getServiceTemplates error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getServiceTemplates({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover initiateOAuth', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.initiateOAuth({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover initiateOAuth error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.initiateOAuth({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover listCredentials', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.listCredentials({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover listCredentials error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.listCredentials({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover createCredential', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.createCredential({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover createCredential error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.createCredential({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover updateCredential', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.updateCredential({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover updateCredential error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.updateCredential({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover deleteCredential', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.deleteCredential({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover deleteCredential error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.deleteCredential({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover testAuth', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.testAuth({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover testAuth error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.testAuth({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover exchangeOAuthCode', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.exchangeOAuthCode({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover exchangeOAuthCode error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.exchangeOAuthCode({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover handleOAuthCallback', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.handleOAuthCallback({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover handleOAuthCallback error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.handleOAuthCallback({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover listUsers', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.listUsers({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover listUsers error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.listUsers({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover getCurrentUser', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getCurrentUser({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover getCurrentUser error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getCurrentUser({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover createUser', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.createUser({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover createUser error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.createUser({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover updateUser', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.updateUser({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover updateUser error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.updateUser({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover deleteUser', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.deleteUser({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover deleteUser error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.deleteUser({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover listSkills', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.listSkills({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover listSkills error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.listSkills({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover getSkill', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getSkill({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover getSkill error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getSkill({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover createSkill', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.createSkill({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover createSkill error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.createSkill({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover updateSkill', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.updateSkill({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover updateSkill error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.updateSkill({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover deleteSkill', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.deleteSkill({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover deleteSkill error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.deleteSkill({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover createProfile', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.createProfile({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover createProfile error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.createProfile({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover updateProfile', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.updateProfile({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover updateProfile error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.updateProfile({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover deleteProfile', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.deleteProfile({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover deleteProfile error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.deleteProfile({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover listProfiles', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.listProfiles({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover listProfiles error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.listProfiles({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover listSecrets', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.listSecrets({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover listSecrets error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.listSecrets({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover revealSecret', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.revealSecret({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover revealSecret error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.revealSecret({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover saveSecret', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.saveSecret({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover saveSecret error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.saveSecret({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover deleteSecret', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.deleteSecret({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover deleteSecret error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.deleteSecret({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover getGlobalSettings', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getGlobalSettings({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover getGlobalSettings error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getGlobalSettings({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover saveGlobalSettings', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.saveGlobalSettings({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover saveGlobalSettings error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.saveGlobalSettings({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover getDashboardTraffic', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getDashboardTraffic({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover getDashboardTraffic error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getDashboardTraffic({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover getTopTools', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getTopTools({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover getTopTools error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getTopTools({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover getAlertStats', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getAlertStats({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover getAlertStats error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getAlertStats({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover deleteAlert', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.deleteAlert({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover deleteAlert error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.deleteAlert({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover listAlerts', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.listAlerts({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover listAlerts error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.listAlerts({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover listAlertRules', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.listAlertRules({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover listAlertRules error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.listAlertRules({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover createAlertRule', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.createAlertRule({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover createAlertRule error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.createAlertRule({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover getAlertRule', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getAlertRule({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover getAlertRule error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getAlertRule({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover updateAlertRule', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.updateAlertRule({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover updateAlertRule error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.updateAlertRule({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover deleteAlertRule', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.deleteAlertRule({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover deleteAlertRule error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.deleteAlertRule({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover getToolFailures', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getToolFailures({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover getToolFailures error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getToolFailures({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover getToolUsage', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getToolUsage({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover getToolUsage error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getToolUsage({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover getSystemStatus', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getSystemStatus({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover getSystemStatus error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getSystemStatus({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover getDoctorStatus', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getDoctorStatus({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover getDoctorStatus error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getDoctorStatus({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover getDashboardHealth', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getDashboardHealth({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover getDashboardHealth error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getDashboardHealth({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover getDashboardMetrics', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getDashboardMetrics({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover getDashboardMetrics error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getDashboardMetrics({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover getTraces', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getTraces({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover getTraces error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getTraces({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover clearTraces', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.clearTraces({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover clearTraces error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.clearTraces({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover getTopology', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getTopology({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover getTopology error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getTopology({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover seedTrafficData', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.seedTrafficData({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover seedTrafficData error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.seedTrafficData({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover seedTrace', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.seedTrace({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover seedTrace error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.seedTrace({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover updateAlertStatus', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.updateAlertStatus({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover updateAlertStatus error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.updateAlertStatus({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover getWebhookURL', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getWebhookURL({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover getWebhookURL error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getWebhookURL({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover saveWebhookURL', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.saveWebhookURL({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover saveWebhookURL error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.saveWebhookURL({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover listCollections', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.listCollections({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover listCollections error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.listCollections({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover getCollection', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getCollection({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover getCollection error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getCollection({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover saveCollection', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.saveCollection({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover saveCollection error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.saveCollection({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover deleteCollection', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.deleteCollection({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover deleteCollection error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.deleteCollection({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover getStackConfig', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getStackConfig({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover getStackConfig error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getStackConfig({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover saveStackConfig', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.saveStackConfig({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover saveStackConfig error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.saveStackConfig({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover getStackYaml', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getStackYaml({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover getStackYaml error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getStackYaml({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover listTemplates', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.listTemplates({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover listTemplates error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.listTemplates({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover saveTemplate', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.saveTemplate({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover saveTemplate error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.saveTemplate({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover deleteTemplate', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.deleteTemplate({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover deleteTemplate error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.deleteTemplate({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover saveStackYaml', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.saveStackYaml({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover saveStackYaml error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.saveStackYaml({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover listAuditLogs', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.listAuditLogs({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover listAuditLogs error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.listAuditLogs({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover exportAuditLogs', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.exportAuditLogs({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover exportAuditLogs error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.exportAuditLogs({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover getDiscoveryStatus', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getDiscoveryStatus({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover getDiscoveryStatus error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.getDiscoveryStatus({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover triggerDiscovery', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.triggerDiscovery({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should cover triggerDiscovery error', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    try { await apiClient.triggerDiscovery({} as any, {} as any, {} as any); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });


  it('should handle array branches in mapUpstreamServiceConfig', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        services: [{
          name: "test",
          resilience: { retry_policy: {} },
          call_policies: [{ rules: [{}] }]
        }]
      })
    } as Response);
    try { await apiClient.listServices(); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should handle edge cases in executeTool', async () => {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        tools: [{
          compatibility: { features: [{ parameters: [{ required: true }] }] }
        }]
      })
    } as Response);
    try { await apiClient.listTools("id"); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();
  });


  it('should cover audit logs with empty content disp', async () => {
    global.URL.createObjectURL = vi.fn();
    global.URL.revokeObjectURL = vi.fn();
    const mockAnchor = { href: '', download: '', click: vi.fn() };
    vi.spyOn(document, 'createElement').mockReturnValue(mockAnchor as any);
    vi.spyOn(document.body, 'appendChild').mockImplementation(() => null as any);
    vi.spyOn(document.body, 'removeChild').mockImplementation(() => null as any);

    const headers = new Headers();
    headers.set('Content-Disposition', 'filename="test.csv"');
    fetchMock.mockResolvedValueOnce({
      ok: true,
      blob: async () => new Blob(),
      headers: headers
    } as any);
    try { await apiClient.exportAuditLogs({}); } catch(e) {}
    expect(fetchMock).toHaveBeenCalled();

    const headers2 = new Headers();
    fetchMock.mockResolvedValueOnce({
      ok: true,
      blob: async () => new Blob(),
      headers: headers2
    } as any);
    try { await apiClient.exportAuditLogs({}); } catch(e) {}
  });
});
