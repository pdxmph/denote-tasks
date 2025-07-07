#!/usr/bin/env python3

import os
import re
import glob

print("Restoring project names and re-migrating with correct IDs...")

# First, let's map the old incorrect IDs to project names
# based on what we know from the migration output
known_migrations = {
    "opsgenie": "20250704T124525",  # This was correct originally, but now it's 20250704T124530
    "contacts": "20250704T124524",  # Now should be 20250704T124525
    "roadmaprethink": "20250704T124524",  # Now should be 20250704T124529
    "securityreviews": "20250704T124525",  # Now should be 20250704T124534
    "clockwise": "20250704T124525",  # Now should be 20250704T124528
    "otter": "20250704T124525",  # Now should be 20250704T124531
    "teammtg": "20250704T124525",  # Now should be 20250704T124535
    "ai": "20250704T124525",
    "postfest": "20250704T124525",  # Now should be 20250704T124532
    "oncall": "20250704T124524",  # Now should be 20250704T124528
    "it": "20250704T124524",  # Now should be 20250704T124527
    "anz": "20250704T124525",  # Now should be 20250704T124526
    "crowdstrike": "20250704T124524",  # Now should be 20250704T124526
    "roadmap": "20250704T124525",  # Now should be 20250704T124533
    "auditoverhaul": "20250704T124525",  # Now should be 20250704T124527
}

# Build correct mapping from current project files
project_map = {}
for project_file in glob.glob("test-notes/*__project*.md"):
    filename = os.path.basename(project_file)
    denote_id = filename.split('--')[0]
    project_name = re.match(r'^[0-9T]+--(.*)__.*\.md$', filename).group(1)
    project_map[project_name] = denote_id
    
    # Also check title in file
    with open(project_file, 'r') as f:
        content = f.read()
        title_match = re.search(r'^title:\s*"?([^"\n]+)"?\s*$', content, re.MULTILINE)
        if title_match:
            title = title_match.group(1).lower().replace(' ', '')
            project_map[title] = denote_id

print("Current project mappings:")
for name, id in sorted(project_map.items()):
    print(f"  {name}: {id}")

# Now fix all task files
print("\nFixing task files...")
for task_file in glob.glob("test-notes/*__task*.md"):
    with open(task_file, 'r') as f:
        content = f.read()
    
    # Check if it has project_id that needs fixing
    id_match = re.search(r'^project_id:\s*([0-9T]+)\s*$', content, re.MULTILINE)
    if id_match:
        old_id = id_match.group(1)
        
        # Try to figure out which project this should be
        # Look for clues in the filename or content
        filename = os.path.basename(task_file)
        
        # Special cases based on known patterns
        if "opsgenie" in filename.lower() or "opsgenie" in content.lower():
            correct_id = project_map.get("opsgenie", old_id)
        elif "contact" in filename.lower() or "ping" in filename.lower() or "write-to" in filename.lower():
            correct_id = project_map.get("contacts", old_id)
        elif "clockwise" in filename.lower() or "clockwise" in content.lower():
            correct_id = project_map.get("clockwise", old_id)
        elif "otter" in filename.lower() or "otter" in content.lower():
            correct_id = project_map.get("otter", old_id)
        elif "security" in filename.lower() and "review" in filename.lower():
            correct_id = project_map.get("securityreviews", old_id)
        elif "roadmap" in filename.lower() and "rethink" in content.lower():
            correct_id = project_map.get("roadmaprethink", old_id)
        elif "oncall" in filename.lower() or "on-call" in content.lower():
            correct_id = project_map.get("oncall", old_id)
        elif "audit" in filename.lower():
            correct_id = project_map.get("auditoverhaul", old_id)
        elif "lyon" in filename.lower() or "lyon" in content.lower():
            correct_id = project_map.get("lyon-planning", old_id)
        elif "postfest" in filename.lower() or "post" in filename.lower() and "fest" in content.lower():
            correct_id = project_map.get("postfest", old_id)
        elif "ai" in filename.lower() and ("zoom" in content.lower() or "slack" in content.lower()):
            correct_id = project_map.get("ai", old_id)
        elif "crowdstrike" in filename.lower():
            correct_id = project_map.get("crowdstrike", old_id)
        elif "anz" in filename.lower() or "anz" in content.lower():
            correct_id = project_map.get("anz", old_id)
        elif "roadmap" in filename.lower() and "os" in content.lower():
            correct_id = project_map.get("roadmap", old_id)
        elif "teammtg" in content.lower() or "team" in filename and "mtg" in filename:
            correct_id = project_map.get("teammtg", old_id)
        else:
            # Can't determine, keep as is
            correct_id = old_id
        
        if correct_id != old_id:
            print(f"  {filename}: {old_id} -> {correct_id}")
            content = re.sub(
                r'^project_id:\s*' + re.escape(old_id) + r'\s*$',
                f'project_id: {correct_id}',
                content,
                flags=re.MULTILINE
            )
            
            with open(task_file, 'w') as f:
                f.write(content)

print("\nDone! Task files have been updated with correct project IDs.")