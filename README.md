# vaultsecrets

Automatically creates .env files from remote vault secrets.

## Warning

Code is designed to work with BiTaksi vault structure. It may not work for others.

## Installation

```sh
go install github.com/ZuluNovember/vaultsecrets@latest
```

* Make sure you have go in your PATH variables

## Usage

1. Change into directory you want to create your .env file.
2. Run ``` vaultsecrets ``` from your terminal
3. You will be prompted to enter your vault server url and vault token. Make sure you enter them correctly
4. You will be prompted to choose stage and repo. `.env` file should be created after this process.

Your credentials will be stored in `$HOME/.vaultconf.ini` file. You can edit this file in place or you can delete the file to fix your credentials. If you delete the file you will be prompted to enter your credentials again.

**Please be careful. You will lose your .env file if you already have that in your directory.**
