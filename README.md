Backup
======

Backup is a backup tool. Use at your own risks!

Installation
------------

Drop the binary for your platform in the *bin* directory of the archive
somewhere in your `PATH` and rename it *backup*.

Usage
-----

Drop the configuration file at the root of your backup medium (such as an USB
key). It should be named *.backup* and look like this:

```yaml
frodo:
  includes:
  - 'doc/perso/**/*'
  excludes:
  - '**/build/**/*'
  - '**/target/**/*'
  - '**/.git/**/*'
  - '**/env/**/*'
  - '**/venv/**/*'
```

This is a map that lists files to include and exclude for given machine.
Paths are relative to the home directory of the user.

To perform a backup, insert and mount your USB key and type:

```bash
$ backup
```

This will list all files while copying them to the backup device.

*In Memory of my Dad*

*Enjoy!*
