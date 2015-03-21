OS X LaunchD Instructions
=========================

 + Edit local.jackpal.seattlesnowman.plist to match your installation path and
   command-line options.

 + Install

   cd seattlesnowman
   go install
   sudo tools/installLaunchDaemon.sh

 + Browse to http://localhost:8080 to view Seattle Snowman running.

 + Uninstall

   sudo tools/uninstallLaunchDaemon.sh
