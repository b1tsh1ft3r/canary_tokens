# Canary Agent

This project will allow the use of "canary tokens" which are dummy files setup to
track lateral movement of an intruder on a system. Accessing the files or modifying 
them will trigger an alert and send information back to a central location for review.
In this case, the central location is a Slack chat channel and the update about
which file and other important data is sent as a message via slack webhook.

The golang program ct.go creates inotify watchers based on data in the ct.conf
file. These watchers can be a directory path or a full path to a specific file.
The file can be owned by anyone or have any content but should be named something
obvious and enticing and put in locations near pieces of software on the system
that might trick users into accessing them.

The content in password.txt and the id_rsa and id_rsa.pub files is completely
randomly generated every time the playbook is run. The data in these files is
NOT used anywhere and is for bait purposes only. Feel free to modify and add or
remove files to be created.

# Instructions
Run the playbook with the following:
ansible-playbook -i inventory.ini ct_deploy.yml

# Gotchas
The ct systemd service wont start unless there are paths (directories or files)
in the /etc/ct.conf file.

ct.conf - doesnt support recursive directories. You need to specific each target explicitly that
          you want to monitor for events.

Monitoring services such as crowdstrike or wazuh NEED to set exclusions for accessing 
these files or you will receive false alerts.

