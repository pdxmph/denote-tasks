#!/usr/bin/env python3
"""
Sync tags from Denote filenames to frontmatter metadata.

This script finds files where tags exist in the filename but not in the
frontmatter metadata, and updates the metadata to include those tags.
"""

import os
import re
import sys
import yaml
from pathlib import Path

def parse_denote_filename(filename):
    """Parse a Denote filename to extract tags."""
    # Pattern: ID--TITLE__TAG1_TAG2.md (double underscore before tags, single between)
    match = re.match(r'^(\d{8}T\d{6})--([^_]+)__(.+?)\.md$', filename)
    if match:
        tags_part = match.group(3)
        if tags_part:
            return tags_part.split('_')
    return []

def read_frontmatter(filepath):
    """Read and parse YAML frontmatter from a file."""
    with open(filepath, 'r') as f:
        content = f.read()
    
    if not content.startswith('---\n'):
        return None, content
    
    # Find the closing ---
    parts = content.split('\n---\n', 2)
    if len(parts) < 2:
        return None, content
    
    try:
        metadata = yaml.safe_load(parts[1])
        body = parts[2] if len(parts) > 2 else ''
        return metadata, body
    except yaml.YAMLError:
        return None, content

def write_frontmatter(filepath, metadata, body):
    """Write updated frontmatter and body back to file."""
    with open(filepath, 'w') as f:
        f.write('---\n')
        yaml.dump(metadata, f, default_flow_style=False, allow_unicode=True)
        f.write('---\n')
        f.write(body)

def sync_tags(directory, dry_run=False):
    """Sync tags from filenames to metadata."""
    updated_count = 0
    
    for filepath in Path(directory).glob('*.md'):
        filename = filepath.name
        file_tags = parse_denote_filename(filename)
        
        if not file_tags:
            continue
        
        metadata, body = read_frontmatter(filepath)
        if metadata is None:
            print(f"Skipping {filename}: No valid frontmatter")
            continue
        
        # Check if it's a task or project
        is_task = 'task' in file_tags
        is_project = 'project' in file_tags
        
        if not (is_task or is_project):
            continue
        
        # Get existing metadata tags
        metadata_tags = metadata.get('tags', [])
        if isinstance(metadata_tags, list) and metadata_tags:
            # Already has metadata tags, skip
            continue
        
        # File has tags in filename but not in metadata
        print(f"\nFile: {filename}")
        print(f"  Filename tags: {file_tags}")
        print(f"  Metadata tags: {metadata_tags}")
        
        if not dry_run:
            # Update metadata with filename tags
            metadata['tags'] = file_tags
            write_frontmatter(filepath, metadata, body)
            print(f"  ✓ Updated metadata tags")
            updated_count += 1
        else:
            print(f"  → Would update metadata tags")
            updated_count += 1
    
    print(f"\n{'Would update' if dry_run else 'Updated'} {updated_count} files")

def main():
    if len(sys.argv) < 2:
        print("Usage: python sync_tags_to_metadata.py <directory> [--dry-run]")
        sys.exit(1)
    
    directory = sys.argv[1]
    dry_run = '--dry-run' in sys.argv
    
    if not os.path.isdir(directory):
        print(f"Error: {directory} is not a directory")
        sys.exit(1)
    
    print(f"Syncing tags in {directory}")
    if dry_run:
        print("DRY RUN - No changes will be made")
    
    sync_tags(directory, dry_run)

if __name__ == '__main__':
    main()