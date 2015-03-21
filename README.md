Seattle Snowman:  A Utility for Managing Internet Access
========================================================

Features
--------

+ Limit Internet access by device.
+ Limit using time and day of week.
+ Easy to give "N Hours" of access to a device.

Requirements
------------

+ An Internet connection.
+ An [EdgeRouter Lite](https://www.ubnt.com/edgemax/edgerouter-lite/) router.
  + You must be comfortable with configuring the Edge Router Lite. It's
    pretty complicated compared to regular home rouers.
+ A computer on your home network that can run a go application and that is
  always on.

Installation
------------

On your server computer:

    $ go get github.com/jackpal/seattlesnowman
    $ seattlesnowman

Then connect to the server (default port 8080) using a web browser.

    $ open http://localhost:8080/

Router Configuration
--------------------

Set up your edge router to have a firewall with internet access controlled by
a firewall group named "SEATTLESNOWMAN_DROP", and a firewall rule to drop
Internet access for anyone in that group. See docs/edgerouter-config for an
example.

Seattle Snowman works by updating the IP addresses included in the
firewall group "SEATTLESNOWMAN_DROP" at the appropriate times. When an IP is in
the group, then access to the Internet is blocked for that device.

Application Configuration
-------------------------

There is an administrator's console at /admin.html that lets you add and
remove devices. One of the fields on the administrator's consol lets you upload a csv file to bulk define your internet devices.

TODO: Include an example csv file.

Persisting Seattle Snowman
--------------------------

On OS X you can (and should) use launchd to ensure Seattle Snowman runs
each time your computer restarts.

(Launchd instructions)[./docs/OS X LaunchD.md]

TODO: Similar instructions for Windows, Linux.

To Do
-----

Concept of controlled, but not automatically getting access. (e.g. Wii, iPad)
Security: Read vs. Read/Write access.
Native apps
Figure out reliable failure modes. Currently if someone saves the router's
state using some other API while a device is not blocked, and then kills
Seattle Snowman, then the device will never get blocked again.

Launchd recipe needs to be able to update and replace a running version.

Need to detect rebooted router (session dies, or just polling.).

Add overrides to force Internet on, off, vacation hours, workday hours.

Developer Hints
---------------

For a speedier development cycle use gin to automatically compile and restart
the server behind a proxy:

  $ go get github.com/codegangsta/gin
  $ gin
  $ open http://localhost:3000
