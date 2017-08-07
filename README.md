Backup
======

Backup is a backup tool.

Installation
------------

Drop the binary for your platform in the *bin* directory of the archive somewhere in your `PATH` and rename it *backup*.

Usage
-----

To perform a backup, type on command line:

```bash
$ backup
```

Example configuration file:

```yaml
gimli:
  includes:
  - /home/media/archives/**/*
  - /home/media/photos/**/*
  excludes:
  - /home/media/archives/**/build/**/*
  - /home/media/archives/**/target/**/*
```

*Enjoy!*
