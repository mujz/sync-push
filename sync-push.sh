#!/usr/bin/env bash
set -e
trap "exit" INT
CYAN='\033[0;36m'
NC='\033[0m'

LOCATIONS_FILE="${HOME}/locations.ini"
IFS=' ' 
# Exclude .git directory and vim tmp files
EXCLUDE='.git|\.sw.*|.\~|4913'

local_dir="$(pwd)"

new_location() {
  read -p "Remote Location (eg. user@host:/path/to/remote/dir): " "remote"
  echo "${local_dir} ${remote}" >> "${LOCATIONS_FILE}"
}

# Get the remote directory from config or from user
location_text=$(grep "${local_dir}" "${LOCATIONS_FILE}") > /dev/null 2>&1 || true
if [[ -z $location_text ]]; then
  new_location
else
  read -ra location <<< "${location_text}" 
  remote=${location[1]}
fi

# set whether to use inotifywait or fswatch depedning on which one exists
# If neither exist, exit with status 1
hash inotifywait 2>/dev/null && use_inotify=true || \
  hash fswatch 2>/dev/null && use_fswatch=true || \
  exit 1

if [ "${use_inotify}" = true ]; then
  # watch recursilvely for file changes and deletions
  # output format: HH:MM:SS - Event:Event filename
  watch="inotifywait -r -m -e close_write -e delete --exclude \"${EXCLUDE}\" --format '%T - %:e %f' --timefmt '%H:%M:%S' \"${local_dir}\""
elif [ ${use_fswatch} = true ]; then
  # watch recursively for the listed events
  # output format: HH:MM:SS absolute/path/to/file Event Event
  watch="fswatch -txr -f '%H:%M:%S' \
    -e $(echo \"${EXCLUDE}\" | sed 's/\|/\ -e\ /g') \
    --event Created --event Updated --event Removed --event Renamed --event MovedFrom --event MovedTo \
    \"${local_dir}\""
fi

# print the output the watch command and sync push
eval "${watch}" | while read notification; do
  echo -e "${CYAN}${notification}${NC}"
  # sync the changes and deletions in the local dir with the remote dir
  # and exclude the gitignored files and .git dir
  # TODO rsync -aiz --delete --exclude-from "${local_dir}/.gitignore" --exclude ".git" ${local_dir} ${remote}
  rsync -aiz --delete --exclude ".git" ${local_dir} ${remote}
done
