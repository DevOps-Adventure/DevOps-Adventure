# ITU Minitwit


---
## Table of Contents
- [Intro](#intro)
- [Setup for Local Development](#setup-for-local-development)

---
### Intro
- Not included yet

---
### Setup for Local Development

### Getting Started

#### Version Control
We use [Github Flow]((https://docs.github.com/de)), so all code changes happen through pull requests.
1. Clone the repo and create a new branch from the release branch.
2. Branches are named with the following convention: `fix/fix-name` or `feature/feature-name` for
more information please visit [Naming conventions for Git](https://medium.com/@abhay.pixolo/naming-conventions-for-git-branches-a-cheatsheet-8549feca2534)
3. For commit messages we use the following convention:
```
feat: add hat wobble
^--^  ^------------^
|     |
|     +-> Summary in present tense.
|
+-------> Type: chore, docs, feat, fix, refactor, style, or test.
```
For more information please visit [Semantic Commit Messages](https://gist.github.com/joshbuchea/6f47e86d2510bce28f8e7f42ae84c716)


#### Pre-requisites
We recommend setting up a virtual environment for local development to 
manage dependencies and to keep your system's Python environment clean. 
Here's how to do it:

1. Open your terminal and navigate to the itu-minitwit` project directory:
   ```
   cd itu-minitwit
   ```
2. Create a virtual environment named venv` within the project folder by executing:`
   ```
   python3 -m venv venv
   ```
3. Activate the virtual environment with the following command:
   ```
   source venv/bin/activate
   ```
   (Note: On Windows, the activation command is venv\Scripts\activate)
4. Install all the necessary packages specified in the requirements.txt` file using pip:
   ```
   pip install -r requirements.txt
   ```
    
We've included a .gitignore file that excludes the venv directory from version control. 
This prevents the environment folder from being pushed to the production repository, 
avoiding potential conflicts and keeping the repository clean.

#### Code Setup
- Not included yet
