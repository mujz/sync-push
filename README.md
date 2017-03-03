# sync-push
Watches and syncs files across two machines over ssh using rsync. 

This is good for those who write code on one machine and run it on another, or if you have an sftp server that you want to keep in sync with a local directory.

## Usage

Navigate to the directory that you want to sync and run:

```bash
sync-push
```

If you're running the command for the first time in this directory, you will be prompted to enter the remote location (ex. mujz@example.com:/home/mujz to put the directory in the home folder of the remote host). Otherwise, that's it!

#### .syncignore

If you have files that you don't want to push to the remote host, just add `.syncignore` file, which uses the same format as [.gitignore](https://git-scm.com/docs/gitignore)

## Additional options:

```
Usage: sync-push [options]
Watch and sync files from current directory to a remote directory

Options:
help			print this message
--delete		delete extraneous files from destination dirs
--version		print version number
```

## Issues

Any issues or suggestions are welcome under the [issues section](https://github.com/mujz/sync-push/issues).

## Contributions

Pull requests would be greatly appreciated.
