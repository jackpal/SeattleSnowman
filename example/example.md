# Example

This directory contains an example [config.json](config.json) configuration file.

You'll need to edit it to adapt it to your network.

The file is in JSON syntax.

    {

Port is the port that Seattle Snowman will run an HTTP server on. Use a web
browser to connect to http://localhost:8080

      "port": 8080,

AddressGroup is the Edgerouter Lite address group name for the "drop list".
This name must match the router configuration.

      "addressGroup": "SEATTLESNOWMAN_DROP",

Database is the path to the sqlite3 database that's used to store the
Seattle Snowman server state.

      "database": "seattlesnowman.db",

RouterAddress is the address of the router's SSH server. It can optionally
have a :PORT if you have configured your router to listen for ssh on a
nonstandard port.

      "routeraddress": "192.168.1.1",

RouterPrivateKeyPath is the path to your router's private ssh key.

      "routerPrivateKeyPath": "/Users/YOURUSERNAME/.ssh/ROUTER_rsa",

Calendar is the calendar of both Internet access times and holidays.
Typically you would update this once a year as new holidays are announced
for your kids school.

      "calendar": {

Location is a [time zone location](http://golang.org/pkg/time/#LoadLocation).

        "location": "America/Los_Angeles",

Hours are "half open", which means
that Internet access starts at starttime and stops at stoptime.

        "schooldayhours": {"starttime": "4:00PM", "endtime": "8:00PM"},
        "vacationhours": {"starttime": "1:00PM", "endtime": "8:00PM"},
        "holidays":[

Holidays are "closed", which means that the holiday start at startday and
ends at the end of endday.

            {"startday": "4/6/15", "endday": "4/10/15"},
            {"startday": "5/22/15", "endday": "5/25/15"}
            ]
        },

        "devices":[

These are the devices to manage. The IP addresses need to be assigned
statically. (Typically this is done using the router's DHCP server.)

            {"ip": "192.168.1.201", "name": "my-first-computer"},
            {"ip": "192.168.1.202", "name": "my-second-computer"},
            {"ip": "192.168.1.203", "name": "my-third-computer"}
        ]
    }
