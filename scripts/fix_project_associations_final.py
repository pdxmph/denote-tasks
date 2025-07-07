#!/usr/bin/env python3

import os
import glob
import re

print("Fixing project associations...")

# Correct project mappings based on filenames and content
task_to_project = {
    # Opsgenie tasks
    "20250624T231749": "20250114T101200",  # opsgenie-sunset
    
    # Contact tasks
    "20250618T052621": "20250114T100300",  # write-to-aliza
    "20250618T052630": "20250114T100300",  # write-to-carla
    "20250618T052646": "20250114T100300",  # ping-trip
    "20250618T052649": "20250114T100300",  # ping-patrick
    "20250618T052651": "20250114T100300",  # ping-nigel
    "20250618T052655": "20250114T100300",  # ping-kara
    "20250619T024051": "20250114T100300",  # ping-kara (another)
    "20250624T185115": "20250114T100300",  # ping-alanna
    "20250619T180240": "20250114T100300",  # check-timeout-jean
    
    # Otter tasks
    "20250618T034403": "20250114T101400",  # find-otter-connections
    "20250624T225025": "20250114T101400",  # otter-integration-disabled
    "20250624T234629": "20250114T101400",  # develop-otter-disconnect
    
    # Clockwise tasks
    "20250624T224951": "20250114T100800",  # serve-non-renewal-notice
    "20250624T234555": "20250114T100800",  # develop-clockwise-project
    
    # Security reviews tasks
    "20250624T232924": "20250114T101700",  # define-slack-gsuite-security
    "20250624T233305": "20250114T101700",  # review-kristen-slack
    "20250624T233331": "20250114T101700",  # review-google-app
    "20250624T233400": "20250114T101700",  # make-project-plan (already fixed)
    "20250626T024756": "20250114T101700",  # communicate-security-process
    "20250704T125817": "20250114T101700",  # review-security-policies
    
    # Roadmap rethink tasks
    "20250702T022607": "20250114T101100",  # conduct-maturity
    "20250702T022626": "20250114T101100",  # develop-mission-statement
    "20250702T022717": "20250114T101100",  # develop-challenges
    "20250704T003443": "20250114T101100",  # pull-out-metrics
    
    # Oncall tasks
    "20250624T234115": "20250114T100900",  # develop-tabletop
    "20250702T022943": "20250114T100900",  # make-pbp-deck
    "20250702T023014": "20250114T100900",  # look-for-decision-tree
    "20250704T044754": "20250114T100900",  # followup-silke
    
    # AI tasks
    "20250618T031606": "20250114T100200",  # zoom-slack-ai
    
    # Crowdstrike tasks
    "20250624T233705": "20250114T100500",  # crowdstrike-reduction
    
    # ANZ tasks
    "20250628T001733": "20250114T100400",  # review-kendal-anz
    
    # Audit tasks
    "20250702T183304": "20250114T100600",  # audit-fixes
    
    # Lyon planning tasks
    "20250627T191225": "20250114T101000",  # planning-for-lyon
    
    # Postfest tasks
    "20250626T175932": "20250114T101500",  # book-accomodations-noman
    "20250626T175946": "20250114T101500",  # book-accomodations-indy
    "20250626T145409": "20250114T101500",  # place-to-stay-catherine
    "20250627T232210": "20250114T101500",  # post-festival-post-fest
    
    # Team meeting tasks
    "20250618T031407": "20250114T101800",  # plan-it-monthly
    "20250618T190434": "20250114T101800",  # it-topical-discussions
    
    # Roadmap tasks
    "20250618T031435": "20250114T101600",  # os-update-planning
    
    # Denote-tasks project
    "20250113T160000": "20250113T152000",  # implement-quick-task
    "20250113T151000": "20250113T152000",  # implement-tui
}

# Fix each task
for task_file in glob.glob("test-notes/*__task*.md"):
    filename = os.path.basename(task_file)
    task_id = filename.split('--')[0]
    
    if task_id in task_to_project:
        correct_project_id = task_to_project[task_id]
        
        with open(task_file, 'r') as f:
            content = f.read()
        
        # Update the project_id
        if 'project_id:' in content:
            # Extract current project_id
            match = re.search(r'^project_id:\s*([0-9T]+)', content, re.MULTILINE)
            if match:
                current_id = match.group(1)
                if current_id != correct_project_id:
                    print(f"Fixing {filename}: {current_id} -> {correct_project_id}")
                    content = re.sub(
                        r'^project_id:\s*' + re.escape(current_id),
                        f'project_id: {correct_project_id}',
                        content,
                        flags=re.MULTILINE
                    )
                    
                    with open(task_file, 'w') as f:
                        f.write(content)
        else:
            # Add project_id if it doesn't exist but should
            print(f"Adding project_id to {filename}")
            # This is more complex - would need to parse frontmatter properly

print("\nDone fixing project associations!")