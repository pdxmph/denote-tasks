#!/usr/bin/env python3

import os
import re
import glob

print("Migrating task files to use project_id instead of project...")

# Create a mapping of project names to Denote IDs
project_map = {}

# Find all project files and extract their names and IDs
for project_file in glob.glob("test-notes/*__project*.md"):
    # Extract the Denote ID from filename
    filename = os.path.basename(project_file)
    denote_id = filename.split('--')[0]
    
    # Extract the project name from the slug (between -- and __)
    match = re.match(r'^[0-9T]+--(.*)__.*\.md$', filename)
    if match:
        project_name = match.group(1)
        project_map[project_name] = denote_id
        
        # Also check the title in the frontmatter
        with open(project_file, 'r') as f:
            content = f.read()
            title_match = re.search(r'^title:\s*"?([^"\n]+)"?\s*$', content, re.MULTILINE)
            if title_match:
                title = title_match.group(1).lower()
                project_map[title] = denote_id
                if title != project_name:
                    print(f"Found project: {project_name} / {title} (ID: {denote_id})")
                else:
                    print(f"Found project: {project_name} (ID: {denote_id})")

# Now process all task files
updated_count = 0
for task_file in glob.glob("test-notes/*__task*.md"):
    with open(task_file, 'r') as f:
        content = f.read()
    
    # Check if file has a project: field
    project_match = re.search(r'^project:\s*"?([^"\n]+)"?\s*$', content, re.MULTILINE)
    if project_match:
        project_name = project_match.group(1).lower()
        
        # Look up the Denote ID
        if project_name in project_map:
            denote_id = project_map[project_name]
            print(f"Updating {os.path.basename(task_file)}: project '{project_name}' -> project_id '{denote_id}'")
            
            # Create a backup
            with open(task_file + '.bak', 'w') as f:
                f.write(content)
            
            # Replace project: with project_id:
            new_content = re.sub(
                r'^project:\s*"?' + re.escape(project_name) + r'"?\s*$',
                f'project_id: {denote_id}',
                content,
                flags=re.MULTILINE | re.IGNORECASE
            )
            
            with open(task_file, 'w') as f:
                f.write(new_content)
            
            updated_count += 1
        else:
            print(f"WARNING: No project found for '{project_name}' in {os.path.basename(task_file)}")

print(f"\nMigration complete! Updated {updated_count} files.")
print("Backup files created with .bak extension")