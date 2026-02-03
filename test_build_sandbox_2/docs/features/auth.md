# Authentication & Users

**Status:** Planned / Partially Implemented

## Goal
Secure access to the MCP Any dashboard and control user permissions. The Authentication system supports multiple providers (OAuth, OIDC) and granular role-based access control.

## Usage Guide

### 1. Login
Navigate to `/login`.
1. Enter your credentials or select an Identity Provider (Google, GitHub, etc.).
2. Click **"Sign In"**.

![Login Screen](screenshots/auth_login.png)

### 2. Manage Users (Admin)
Navigate to `/users` (accessible only to Admins).
- **List**: View all registered users and their roles.
- **Invite**: Send invitation links to new team members.

![Users List](screenshots/auth_users_list.png)
