# Git Repo Cleaner
Optimizes the size of the .git directory by removing all of the files that are unnecessarily-still-saved as part of the git history.

Unnecessarily-still-saved files are files that fall under one of the following categories.
- **Files ignored by git.** Sometimes we ask git to track some files and then add them to .gitignore, these files will then not be tracked anymore but will still be saved in history so that you can find them when you checkout an old commit.
- **Files that are not found in the local directory of the repo** after checking out to the main branch (e.g. master or main). Sometimes we add files that we would later not need and delete. Even after deleting these files, they will still be saved as part of the git history.

## Usage
```
build clean-repo.go && ./clean-repo -repo-absolute-path=<absolute-path> -main-branch-name=<branch>
```
For example:
```
build clean-repo.go && ./clean-repo -repo-absolute-path=/Users/omar/Documents/git-repo-cleaner -main-branch-name=master
```
