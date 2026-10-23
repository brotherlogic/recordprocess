#!/bin/bash

# Ensure the 'recordprocess' session exists
if ! tmux has-session -t recordprocess 2>/dev/null; then
  # Create a new session named 'recordprocess', detached
  cd /workspaces/recordprocess
  tmux new-session -d -s recordprocess
 fi
