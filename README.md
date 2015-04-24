Seattle Snowman:  A Utility for Managing Internet Access
========================================================

Copyright (C) 2015 John Howard Palevich. All Rights Reserved.

Introduction
------------

This is an app I wrote to help me manage Internet access for my home network.

I've published it for the enjoyment and inspiration of other parents.

Features
--------

+ Limit Internet access by device.
+ Limit using time and day of week.
+ Knows about school days vs. vacation days.
+ Knows about school holidays!
+ Easy UI for giving "N Hours" of access to a given device.

Requirements
------------

+ An Internet connection.
+ An [EdgeRouter Lite](https://www.ubnt.com/edgemax/edgerouter-lite/) router.
  + You must be comfortable with configuring the Edge Router Lite. It's
    pretty complicated compared to regular home routers.
+ A computer on your home network that can run a go application and that is
  always on. (Tested on OS X, probably works on Windows and Linux as well.)

Installation
------------

[Install Go](http://golang.org/doc/install).

Use the Go tool to download and build the Seattle Snowman app:

    $ go get github.com/jackpal/SeattleSnowman
    $ go install


Router Configuration
--------------------

You will need to configure your router to support the Seattle Snowman app.

[EdgeRoute Lite configuration](edgerouterdoc/edgerouter.md) documentation.

Application Configuration
-------------------------

You will need to configure the Seattle Snowman application to know about your
computer network.

When Seattle Snowman launches, it reads a configuration file. By default it
reads the file config.json, but this can be overridden by using the --config
command line flag.

The format of the configuration file is documented [here](example/example.md).

You will need to write your own config.json file that matches your network.

Start the app
-------------

Once you've configured your router and the application, you can start the
application from the command line:

    $ SeattleSnowman

Then use a web browser to connect to the server (default port 8080).

    $ open http://localhost:8080/

Using the Web UI
----------------

The main URL is http://localhost:8080/devices.html

It displays a list of device names along with "+" buttons.

Click on the "+" button to add an hour of Internet access to the device. Each
time you click on the "+" button another hour will be added.

For each device that can currently access the Internet, the "end time" is
displayed, as well as a "-" button that can be clicked to subtract an hour of
Internet time.

Pro Tip: You can save a bookmark on an Android or iOS device for easy access.

Admin Console
-------------

There is an administrator's console at http://localhost:8080/admin.html that
lets you poke at the internals of the application using a series of forms.


Launching Seattle Snowman When your Computer Starts
---------------------------------------------------

On OS X you can (and should) use launchd to ensure Seattle Snowman runs
each time your computer restarts.

[Launchd instructions](./docs/OS X LaunchD.md)

TODO: Provide instructions for Windows, Linux.

To Do
-----

+ Document how to add support for other routers.

+ Add support for other routers. (Hopefully provided by people who port the
app to work with their router.)

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

FAQ
---

Q: Why is it named "Seattle Snowman"?

A: Because I wrote it in the Seattle winter. Also, it's so complicated that
"snow man" can understand it.

Q: Do you really think this will be useful to anyone beside you?

A: Maybe not at the moment, but perhaps in the future if more types of
routers are supported.

Q: Won't your kids just use the app to grant themselves Internet time?

A: Yes, probably. Once they do that I'll add "Basic Auth".

Q: Won't your kids just manually change their device's IP addresses?

A: Well, perhaps. The Edgerouter has some advanced router configuration
tools  that make it harder to do this. I guess I'll investigate that if/when it
becomes a problem.

Q: Won't your kids just use their phones?

A: Yes. At the moment only one of my kids is old enough for a phone, but in
a year or two they'll all have them. The only thing I can think of is to collect
the phones every night and give them back in the morning. Jailer dad.

Q: What about IPv6?

A: Unfortunately this app does nothing to prevent access by IPv6 addresses.
In theory it could be possible to add support for DHCPv6, but the snag there
is that Android devices do not support DHCPv6. I guess the work-around is to
disable IPv6 networking.
