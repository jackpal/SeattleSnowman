# EdgeRouter Lite documentation

Seattle Snowman works by updating the IP addresses included in the
firewall group "SEATTLESNOWMAN_DROP" at the appropriate times. When an IP is in
the group, then access to the Internet is blocked for that device.

# Configuring the EdgeRouter Lite

I'm sorry to say that this is a complicated topic. I wish I knew how to
make it easier.

You will need to configure the EdgeRouter Lite as follows:


 + Enable SSH access with a private/public key. See the EdgeRouter wiki:
   [Access Using SSH](http://wiki.ubnt.com/Access_Using_SSH).

 + Configure the EdgeRouter's DHCP server to assign static IP addresses to all
   the devices you want to manage. (Typically you will do this by collecting
   the MAC addresses of all the devices you want to manage, then tell the
   EdgeRouter's DHCP server to assign static IP addresses based on the MAC
   addresses.)

 + Create a firewall address group "SEATTLESNOWMAN_DROP".

 + Create a WAN_OUT firewall rule to drop Internet access to members of the
   "SEATTLESNOWMAN_DROP" internet group.

If you have trouble, consult the
[EdgeMAX Forums](http://community.ubnt.com/t5/EdgeMAX/bd-p/EdgeMAX).
Remember that EdgeRouter OS is based on the Vyatta OS, so
Vyatta OS documentation, tutorials and examples often apply to the EdgeRouter.

Here is a part of a sample EdgeRouter configuration file showing the entries
you will need to set up.

    firewall {
        group {
             address-group SEATTLESNOWMAN_DROP {
                 description "Seattle Snowman managed devices."
             }
         }
         name WAN_OUT {
             default-action accept
             rule 1000 {
                 action drop
                 description "Limit group SEATTLESNOWMAN_DROP."
                 source {
                     group {
                         address-group SEATTLESNOWMAN_DROP
                     }
                 }
             }
         }
    }

    service {
         dhcp-server {
             shared-network-name LAN1 {
                 subnet 10.10.1.0/24 {
                     static-mapping my-first-computer {
                         ip-address 192.168.1.201
                         mac-address 12:34:56:78:9a:bc
                     }
                 }
             }
         }
     }

Here's my [complete sanitized configuration file](sanitized.config).
