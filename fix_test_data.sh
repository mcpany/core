sed -i 's/requestContext?: APIRequestContext/_requestContext?: APIRequestContext/g' ui/tests/e2e/test-data.ts
sed -i 's/requestContext: APIRequestContext/_requestContext: APIRequestContext/g' ui/tests/e2e/test-data.ts
sed -i 's/requestContext ||/_requestContext ||/g' ui/tests/e2e/test-data.ts
sed -i 's/username: string/_username: string/g' ui/tests/e2e/test-data.ts
sed -i 's/catch (e) {/catch (_e) {/g' ui/tests/e2e/test-data.ts
