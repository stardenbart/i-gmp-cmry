import os, re

def fix_where_clause(m):
    inner = m.group(1)
    
    # We want to replace identifiers that start with an uppercase letter
    # e.g. Username, RoleID, UserID, LOWER(RoleName), rp.RoleID
    def repl_identifier(x):
        word = x.group(1)
        # Skip SQL keywords
        if word in ['LIKE', 'AND', 'OR', 'IN', 'LOWER', 'SELECT', 'FROM', 'WHERE', 'EXTRACT', 'YEAR', 'MONTH', 'AS', 'ON', 'JOIN']:
            return word
        return f'\\"{word}\\"'
    
    inner2 = re.sub(r'\b([A-Z][a-zA-Z0-9_]*)\b', repl_identifier, inner)
    return f'Where("{inner2}"'

modified_files = 0
for root, dirs, files in os.walk('backend/internal'):
    for file in files:
        if file.endswith('.go'):
            path = os.path.join(root, file)
            with open(path, 'r') as f:
                content = f.read()
            
            content2 = re.sub(r'Where\("([^"]+)"', fix_where_clause, content)
            
            if content != content2:
                with open(path, 'w') as f:
                    f.write(content2)
                modified_files += 1

print(f"Modified {modified_files} files")
