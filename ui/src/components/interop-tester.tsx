import React, { useState } from 'react';
import { fetchWithAuth } from '../lib/client'; // Use the main fetch function

export const InteropTester: React.FC = () => {
  const [framework, setFramework] = useState('CrewAI');
  const [intent, setIntent] = useState('task_delegation');
  const [payloadRole, setPayloadRole] = useState('data_analyst');
  const [result, setResult] = useState<string>('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const response = await fetchWithAuth('/api/v1/interop/task', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          framework,
          intent,
          payload: { role: payloadRole },
        }),
      });
      const data = await response.json();
      setResult(JSON.stringify(data, null, 2));
    } catch (error) {
      setResult('Error submitting task: ' + String(error));
    }
  };

  return (
    <div className="p-4" id="interop-tester">
      <h2 className="text-2xl mb-4">Interop Tester</h2>
      <form onSubmit={handleSubmit} className="flex flex-col gap-4 max-w-md">
        <div>
          <label className="block mb-1">Framework</label>
          <input
            id="framework-input"
            type="text"
            value={framework}
            onChange={(e) => setFramework(e.target.value)}
            className="border p-2 w-full"
          />
        </div>
        <div>
          <label className="block mb-1">Intent</label>
          <input
            id="intent-input"
            type="text"
            value={intent}
            onChange={(e) => setIntent(e.target.value)}
            className="border p-2 w-full"
          />
        </div>
        <div>
          <label className="block mb-1">Payload Role</label>
          <input
            id="payload-role-input"
            type="text"
            value={payloadRole}
            onChange={(e) => setPayloadRole(e.target.value)}
            className="border p-2 w-full"
          />
        </div>
        <button id="submit-interop-btn" type="submit" className="bg-blue-500 text-white p-2 rounded">
          Submit Task
        </button>
      </form>
      {result && (
        <div className="mt-4">
          <h3 className="text-xl mb-2">Result:</h3>
          <pre id="interop-result" className="bg-gray-100 p-4 border">{result}</pre>
        </div>
      )}
    </div>
  );
};
