#!/bin/bash

# Migration script to convert project: field to project_id: field
# with proper Denote IDs

echo "Migrating task files to use project_id instead of project..."

# Create a mapping of project names to Denote IDs by scanning project files
declare -A project_map

# Find all project files and extract their names and IDs
for project_file in test-notes/*__project*.md; do
    if [[ -f "$project_file" ]]; then
        # Extract the Denote ID from filename
        filename=$(basename "$project_file")
        denote_id=$(echo "$filename" | cut -d'-' -f1)
        
        # Extract the project name from the slug (between -- and __)
        project_name=$(echo "$filename" | sed -E 's/^[0-9T]+--(.*)__.*\.md$/\1/')
        
        # Also check the title in the frontmatter for the canonical name
        title=$(grep -A1 "^title:" "$project_file" | tail -1 | sed 's/^title: "\(.*\)"$/\1/' | tr '[:upper:]' '[:lower:]')
        
        echo "Found project: $project_name (ID: $denote_id)"
        project_map["$project_name"]="$denote_id"
        
        # Also map the title if different
        if [[ -n "$title" && "$title" != "$project_name" ]]; then
            project_map["$title"]="$denote_id"
        fi
    fi
done

# Now process all task files
for task_file in test-notes/*__task*.md; do
    if [[ -f "$task_file" ]]; then
        # Check if file has a project: field
        if grep -q "^project:" "$task_file"; then
            # Extract the project name
            project_name=$(grep "^project:" "$task_file" | sed 's/^project: *"\?\([^"]*\)"\?$/\1/' | tr '[:upper:]' '[:lower:]')
            
            # Look up the Denote ID
            if [[ -n "${project_map[$project_name]}" ]]; then
                denote_id="${project_map[$project_name]}"
                echo "Updating $task_file: project '$project_name' -> project_id '$denote_id'"
                
                # Create a backup
                cp "$task_file" "$task_file.bak"
                
                # Replace project: with project_id:
                sed -i '' "s/^project: *\"*${project_name}\"*/project_id: ${denote_id}/" "$task_file"
            else
                echo "WARNING: No project found for '$project_name' in $task_file"
            fi
        fi
    fi
done

echo "Migration complete!"
echo "Backup files created with .bak extension"