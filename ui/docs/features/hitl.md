# HITL Approval Interface

The HITL (Human-in-the-Loop) Approval Interface provides a secure, real-time notification and approval flow for actions intercepted by the HITL middleware.

## Overview

When the backend suspends an agent's execution for a high-risk task (e.g., dropping a database table or modifying production infrastructure), a notification is instantly pushed to the HITL interface.

## Usage

1. Navigate to the **Approvals** or **HITL** dashboard.
2. Review the pending actions, including the agent's intent, the tool requested, and the provided arguments.
3. Use the interface to "Approve" or "Deny" the action.
4. If MFA is configured, the system will prompt for additional authentication before releasing the action.

## Security

The interface leverages multi-factor attestation to verify the administrator's identity before completing the approval process.
