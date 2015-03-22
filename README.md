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

When Seattle Snowman launches, it reads a configuration file. (By default it
reads the file config.json, but this can be overridden by using the --config
command line flag.)

The format of the configuration file is documented [here](example/example.md).

Admin Console
-------------

There is an administrator's console at /admin.html that lets you add and
remove devices. One of the fields on the administrator's consol lets you upload a csv file to bulk define your internet devices.

TODO: Include an example csv file.

Launching Seattle Snowman When your Computer Starts
---------------------------------------------------

On OS X you can (and should) use launchd to ensure Seattle Snowman runs
each time your computer restarts.

(Launchd instructions)[./docs/OS X LaunchD.md]

TODO: Provide instructions for Windows, Linux.

To Do
-----

+ Add concept of devices that controlled, but not automatically getting access.
(For game consoles.)

+ Add security: Read vs. Read/Write access.

+ Write native apps.

+ Figure out reliable failure modes. Currently if someone saves the router's
state using some other API while a device is not blocked, and then kills
Seattle Snowman, then the device will never get blocked again.

+ Need to detect and reconnect to rebooted router.

+ Add overrides to force Internet on, off, vacation hours, workday hours.

Developer Tips
--------------

For a speedier development cycle use [gin](https://github.com/codegangsta/gin)
to automatically compile and restart the server behind a proxy:

  $ go get github.com/codegangsta/gin
  $ gin
  $ open http://localhost:3000
