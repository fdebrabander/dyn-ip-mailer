# dyn-ip-mailer
Tiny application written in Go that checks if your external Internet IP has
changed. Sends an e-mail to inform you about your new IP address.

Uses [www.ipify.org](http://www.ipify.org) to lookup the IP address.

## usage
    go install github.com/fdebrabander/dyn-ip-mailer

Copy the example configuration file .dyn-ip-mailer.yaml to $HOME, then
configure the correct SMTP settings.

Add an entry to the cron with *crontab -e* to periodically check if your
IP changed. Please be nice to ipify.org! Replace 17 with some random
minute.

    17 */4 * * * dyn-ip-mailer > /dev/null 2>> $HOME/.dyn-ip-errors

