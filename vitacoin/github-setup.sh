#!/bin/bash

###############################################################################
###                    GitHub Repository Setup Script                       ###
###############################################################################

echo "🚀 Setting up GitHub remote for VitaCoin..."
echo ""

# Check if remote already exists
if git remote | grep -q "origin"; then
    echo "⚠️  Remote 'origin' already exists:"
    git remote -v
    echo ""
    read -p "Do you want to remove it and add a new one? (y/N): " confirm
    if [[ $confirm == [yY] ]]; then
        git remote remove origin
        echo "✅ Removed existing remote"
    else
        echo "❌ Keeping existing remote. Exiting."
        exit 0
    fi
fi

# Ask for GitHub username
read -p "Enter your GitHub username: " github_user

if [ -z "$github_user" ]; then
    echo "❌ GitHub username cannot be empty"
    exit 1
fi

# Add remote
echo ""
echo "📡 Adding remote repository..."
git remote add origin "git@github.com:${github_user}/vitacoin.git"

if [ $? -eq 0 ]; then
    echo "✅ Remote added successfully!"
    echo ""
    git remote -v
    echo ""
    
    # Ask if user wants to push
    read -p "Do you want to push to GitHub now? (y/N): " push_now
    if [[ $push_now == [yY] ]]; then
        echo ""
        echo "📤 Pushing to GitHub..."
        git push -u origin main
        
        if [ $? -eq 0 ]; then
            echo ""
            echo "🎉 Successfully pushed to GitHub!"
            echo "🔗 Repository URL: https://github.com/${github_user}/vitacoin"
        else
            echo ""
            echo "❌ Push failed. Make sure:"
            echo "   1. You've created the repository on GitHub"
            echo "   2. Your SSH key is configured correctly"
            echo "   3. You have write access to the repository"
        fi
    else
        echo ""
        echo "📝 To push later, run: git push -u origin main"
    fi
else
    echo "❌ Failed to add remote"
    exit 1
fi
