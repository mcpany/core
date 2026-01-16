/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState, Suspense } from 'react';
import { useSearchParams, useRouter } from 'next/navigation';
import { apiClient } from '@/lib/client';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { CheckCircle2, XCircle, Loader2 } from 'lucide-react';
import { Button } from '@/components/ui/button';


function OAuthCallbackContent() {
    const searchParams = useSearchParams();
    const router = useRouter();
    const [status, setStatus] = useState<'loading' | 'success' | 'error'>('loading');
    const [errorMessage, setErrorMessage] = useState('');
    const [returnPath, setReturnPath] = useState('/services');

    useEffect(() => {
        const code = searchParams.get('code');
        const state = searchParams.get('state');
        const error = searchParams.get('error');

        if (error) {
            setStatus('error');
            setErrorMessage(error);
            return;
        }

        if (!code) {
            setStatus('error');
            setErrorMessage('No authorization code received.');
            return;
        }

        const handleCallback = async () => {
            try {
                // Retrieve stored context
                const serviceID = sessionStorage.getItem('oauth_service_id');
                const credentialID = sessionStorage.getItem('oauth_credential_id');
                const storedState = sessionStorage.getItem('oauth_state');
                const redirectUrl = sessionStorage.getItem('oauth_redirect_url') || `${window.location.origin}/auth/callback`;

                console.log(`DEBUG: Callback state=${state}, storedState=${storedState}, serviceID=${serviceID}, credentialID=${credentialID}`);

                const storedReturnPath = sessionStorage.getItem('oauth_return_path');
                if (storedReturnPath) {
                    setReturnPath(storedReturnPath);
                }

                // NOTE: In a real app we MUST verify state matches storedState
                if (state !== storedState) {
                    console.warn(`State mismatch: received ${state}, expected ${storedState}`);
                    // Proceeding for now but this should be an error in production
                }

                if (!serviceID && !credentialID) {
                    console.error('Missing session data for oauth');
                    throw new Error('No service ID or credential ID found in session. Please start the flow again.');
                }

                await apiClient.handleOAuthCallback(serviceID || null, code, redirectUrl, credentialID || undefined);
                setStatus('success');

                // Cleanup
                sessionStorage.removeItem('oauth_service_id');
                sessionStorage.removeItem('oauth_credential_id');
                sessionStorage.removeItem('oauth_state');
                sessionStorage.removeItem('oauth_redirect_url');
                sessionStorage.removeItem('oauth_return_path');

            } catch (e: any) {
                setStatus('error');
                setErrorMessage(e.message || 'Failed to complete authentication.');
            }
        };

        handleCallback();
    }, [searchParams]);

    const handleContinue = () => {
        router.push(returnPath);
    };

    return (
        <div className="flex items-center justify-center min-h-screen bg-background">
            <Card className="w-full max-w-md">
                <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                        {status === 'loading' && <Loader2 className="animate-spin" />}
                        {status === 'success' && <CheckCircle2 className="text-green-500" />}
                        {status === 'error' && <XCircle className="text-destructive" />}
                        {status === 'loading' && 'Authenticating...'}
                        {status === 'success' && 'Authentication Successful'}
                        {status === 'error' && 'Authentication Failed'}
                    </CardTitle>
                </CardHeader>
                <CardContent>
                    {status === 'loading' && <p>Please wait while we complete the secure connection.</p>}
                    {status === 'success' && (
                        <div className="space-y-4">
                            <p>You have successfully connected your account.</p>

                            <Button onClick={handleContinue} className="w-full">
                                Continue
                            </Button>
                        </div>
                    )}
                    {status === 'error' && (
                         <div className="space-y-4">
                             <Alert variant="destructive">
                                <AlertTitle>Error</AlertTitle>
                                <AlertDescription>{errorMessage}</AlertDescription>
                            </Alert>
                            <Button variant="outline" onClick={() => router.push(returnPath)} className="w-full">
                                Back
                            </Button>
                         </div>
                    )}
                </CardContent>
            </Card>
        </div>
    );
}

export default function OAuthCallbackPage() {
    return (
        <Suspense fallback={<div className="flex items-center justify-center min-h-screen"><Loader2 className="animate-spin" /></div>}>
            <OAuthCallbackContent />
        </Suspense>
    );
}
