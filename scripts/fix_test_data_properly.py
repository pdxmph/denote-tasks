#!/usr/bin/env python3

import os
import glob
import re
import yaml

print("Fixing test data to match Denote Task Spec v3.0.0...")

# Counter for the unified index_id
current_index = 1

# Track project Denote IDs for task associations
project_denote_ids = {}

def fix_file(filepath):
    """Fix a single file to match spec v3.0.0"""
    global current_index
    
    with open(filepath, 'r') as f:
        content = f.read()
    
    # Extract frontmatter and body
    if content.startswith('---\n'):
        try:
            # Split on --- delimiters
            parts = content.split('\n---\n', 2)
            if len(parts) >= 2:
                fm_text = parts[1]
                body = parts[2] if len(parts) > 2 else ""
            else:
                print(f"  WARNING: Could not parse frontmatter in {filepath}")
                return False
                
            # Parse frontmatter
            fm = yaml.safe_load(fm_text)
            if not fm:
                fm = {}
            
            # Get filename info
            filename = os.path.basename(filepath)
            denote_id = filename.split('--')[0]
            is_task = '__task' in filename
            is_project = '__project' in filename
            
            # Remove incorrect fields
            if 'id' in fm:
                del fm['id']
            if 'identifier' in fm:
                del fm['identifier']
            
            # Fix the index_id field
            if is_task:
                # Convert task_id to index_id
                if 'task_id' in fm:
                    fm['index_id'] = current_index
                    current_index += 1
                    del fm['task_id']
                elif 'index_id' not in fm:
                    fm['index_id'] = current_index
                    current_index += 1
            elif is_project:
                # Convert project_id to index_id
                if 'project_id' in fm and isinstance(fm['project_id'], int):
                    fm['index_id'] = current_index
                    current_index += 1
                    del fm['project_id']
                elif 'index_id' not in fm:
                    fm['index_id'] = current_index
                    current_index += 1
                
                # Track project Denote ID for later
                project_name = filename.split('--')[1].split('__')[0]
                project_denote_ids[project_name] = denote_id
                
                # Also track by title if available
                if 'title' in fm:
                    title_key = fm['title'].lower().replace(' ', '-')
                    project_denote_ids[title_key] = denote_id
            
            # Ensure required fields
            if 'title' not in fm and (is_task or is_project):
                # Generate title from filename
                title_slug = filename.split('--')[1].split('__')[0]
                title = title_slug.replace('-', ' ').title()
                fm['title'] = title
            
            # Fix date format if needed
            if 'date' in fm:
                del fm['date']  # This field shouldn't exist
            
            # Write back the fixed frontmatter
            new_content = "---\n"
            # Write fields in a sensible order
            field_order = ['title', 'index_id', 'type', 'status', 'priority', 
                          'due_date', 'start_date', 'estimate', 'project_id', 
                          'area', 'assignee', 'tags']
            
            for field in field_order:
                if field in fm:
                    if isinstance(fm[field], list):
                        new_content += f"{field}: {fm[field]}\n"
                    elif isinstance(fm[field], str) and ' ' in fm[field]:
                        new_content += f'{field}: "{fm[field]}"\n'
                    else:
                        new_content += f"{field}: {fm[field]}\n"
            
            # Add any remaining fields not in our order
            for field, value in fm.items():
                if field not in field_order:
                    if isinstance(value, list):
                        new_content += f"{field}: {value}\n"
                    elif isinstance(value, str) and ' ' in value:
                        new_content += f'{field}: "{value}"\n'
                    else:
                        new_content += f"{field}: {value}\n"
            
            new_content += "---\n"
            new_content += body
            
            # Write the file
            with open(filepath, 'w') as f:
                f.write(new_content)
            
            return True
            
        except yaml.YAMLError as e:
            print(f"  ERROR: YAML error in {filepath}: {e}")
            return False
    else:
        print(f"  WARNING: No frontmatter found in {filepath}")
        return False

# First pass: Fix all files and collect project info
print("\nFirst pass: Fixing files and collecting project info...")
for filepath in sorted(glob.glob("test-notes/*.md")):
    filename = os.path.basename(filepath)
    if '__task' in filename or '__project' in filename:
        print(f"  Fixing {filename}")
        fix_file(filepath)

# Second pass: Fix project_id references in tasks
print("\nSecond pass: Fixing project_id references in tasks...")
for filepath in glob.glob("test-notes/*__task*.md"):
    with open(filepath, 'r') as f:
        content = f.read()
    
    # Check if it has a project_id that needs fixing
    if 'project_id:' in content:
        # Extract frontmatter
        if content.startswith('---\n'):
            parts = content.split('\n---\n', 2)
            if len(parts) >= 2:
                fm_text = parts[1]
                body = parts[2] if len(parts) > 2 else ""
                fm = yaml.safe_load(fm_text)
                
                if 'project_id' in fm and isinstance(fm['project_id'], str):
                    # Check if it's already a valid Denote ID
                    if not re.match(r'^\d{8}T\d{6}$', fm['project_id']):
                        # Try to find the right project
                        filename = os.path.basename(filepath)
                        
                        # Look for project name in various places
                        for project_name, denote_id in project_denote_ids.items():
                            if project_name in filename.lower() or project_name in content.lower():
                                print(f"    Updating {filename}: project_id -> {denote_id}")
                                fm['project_id'] = denote_id
                                
                                # Rewrite the file
                                new_content = "---\n"
                                for key, value in fm.items():
                                    if isinstance(value, list):
                                        new_content += f"{key}: {value}\n"
                                    elif isinstance(value, str) and ' ' in value:
                                        new_content += f'{key}: "{value}"\n'
                                    else:
                                        new_content += f"{key}: {value}\n"
                                new_content += "---\n"
                                new_content += body
                                
                                with open(filepath, 'w') as f:
                                    f.write(new_content)
                                break

# Create/update the counter file
counter_data = {
    "next_index": current_index,
    "spec_version": "3.0.0"
}

counter_path = "test-notes/.denote-task-counter.json"
import json
with open(counter_path, 'w') as f:
    json.dump(counter_data, f, indent=2)

print(f"\nCreated counter file with next_index: {current_index}")
print("\nDone! Test data should now conform to spec v3.0.0")