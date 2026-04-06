1. **Understand requirements**: Review and add docstrings to every public function, method, class, and exported constant in the codebase (Go and TypeScript/JS) with specific sections: `Summary:`, `Parameters:`, `Returns:`, `Throws/Errors:`, `Side Effects:`. Then rewrite `README.md` to match the "Gold Standard" (Elevator Pitch, Architecture, Getting Started, Development, Configuration).

2. **Automate Go docstring addition**: I will create a python script that utilizes the `ast` module or simple regex parsing to iterate over all Go files and correctly inject missing sections for exported functions, types, constants, and variables. The script will look for missing sections and append them to existing docstrings or create new ones.

3. **Automate TypeScript docstring addition**: Similarly, I will create a python script that iterates over all TypeScript/JS files and injects missing sections for exported functions, classes, interfaces, types, constants, and enums.

4. **Run the scripts and apply fixes**: Execute the scripts and ensure the docstrings are correctly added to all public symbols.

5. **Fix the README.md**: Rewrite the `README.md` based on the specified structure. It currently has a similar structure, but I will review it to ensure it perfectly matches the requirements.

6. **Pre commit step**: Use `pre_commit_instructions` tool to complete testing and verifications.

7. **Verify**: Run `make lint` and `make test`. Ensure no errors.

8. **Submit**: Submit the changes.
