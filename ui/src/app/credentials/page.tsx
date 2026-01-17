/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { CredentialList } from "@/components/credentials/credential-list"
import { Separator } from "@/components/ui/separator"

/**
 * CredentialsPage displays a list of credentials.
 *
 * @returns {JSX.Element} The rendered Credentials page.
 */
export default function CredentialsPage() {
  return (
    <div className="h-full flex-1 flex-col space-y-8 p-8 md:flex">
      <div className="flex items-center justify-between space-y-2">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">External Authentication</h2>
          <p className="text-muted-foreground">
            Manage reusable authentication credentials for upstream services.
          </p>
        </div>
      </div>
      <Separator />
      <CredentialList />
    </div>
  )
}
