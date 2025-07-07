#!/usr/bin/env python3

import os
import glob
import re
from datetime import datetime, timedelta

print("Fixing duplicate project IDs...")

# Group project files by timestamp
timestamp_groups = {}
for project_file in sorted(glob.glob("test-notes/*__project*.md")):
    filename = os.path.basename(project_file)
    timestamp = filename.split('--')[0]
    
    if timestamp not in timestamp_groups:
        timestamp_groups[timestamp] = []
    timestamp_groups[timestamp].append(project_file)

# Show the problem
print("\nDuplicate timestamps found:")
for timestamp, files in timestamp_groups.items():
    if len(files) > 1:
        print(f"\nTimestamp {timestamp} used by {len(files)} projects:")
        for f in files:
            print(f"  - {os.path.basename(f)}")

# Fix duplicates by incrementing seconds
print("\n\nFixing duplicates by assigning unique timestamps...")
for timestamp, files in timestamp_groups.items():
    if len(files) > 1:
        # Parse the timestamp
        dt = datetime.strptime(timestamp, "%Y%m%dT%H%M%S")
        
        # Keep the first file as-is, increment others
        for i, old_path in enumerate(files[1:], 1):
            # Increment by i seconds
            new_dt = dt + timedelta(seconds=i)
            new_timestamp = new_dt.strftime("%Y%m%dT%H%M%S")
            
            # Build new filename
            old_filename = os.path.basename(old_path)
            new_filename = old_filename.replace(timestamp, new_timestamp)
            new_path = os.path.join(os.path.dirname(old_path), new_filename)
            
            print(f"  Renaming: {old_filename}")
            print(f"        to: {new_filename}")
            
            # Rename the file
            os.rename(old_path, new_path)
            
            # Update the ID inside the file
            with open(new_path, 'r') as f:
                content = f.read()
            
            # Update the id field in frontmatter
            content = re.sub(
                r'^id:\s*"?' + re.escape(timestamp) + r'"?\s*$',
                f'id: "{new_timestamp}"',
                content,
                flags=re.MULTILINE
            )
            
            with open(new_path, 'w') as f:
                f.write(content)

print("\nProject IDs fixed! Now you'll need to re-run the migration to update task project_id fields.")