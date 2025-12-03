import os

def test_readme_for_mcp_any_cli():
    """
    This test checks the README.md file for any occurrences of the string "mcp-any-cli".
    """
    readme_path = os.path.join(os.path.dirname(__file__), '..', 'README.md')
    with open(readme_path, 'r') as f:
        content = f.read()
    assert "mcp-any-cli" not in content, "The string 'mcp-any-cli' should not be present in the README.md file."
