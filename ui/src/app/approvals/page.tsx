import React from 'react';

export default function ApprovalsPage() {
  return (
    <div className="p-8">
      <h1 className="text-2xl font-bold mb-4">HITL Approval Interface</h1>
      <p className="text-gray-600 mb-8">
        Secure, real-time notification and approval flow for actions intercepted by the HITL middleware.
      </p>

      <div className="border rounded-lg p-6 bg-white shadow-sm">
        <h2 className="text-xl font-semibold mb-4">Pending Approvals</h2>
        <div className="text-center p-8 text-gray-500 border border-dashed rounded bg-gray-50">
          No pending actions requiring human approval at this time.
        </div>
      </div>
    </div>
  );
}
