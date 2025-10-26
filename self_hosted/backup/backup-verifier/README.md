# Backup Verifier

A JavaScript script that monitors backup folders and sends email notifications when backups are missing or empty.

## Features

- ✅ Checks if backup folder exists for current date
- ✅ Verifies folder is not empty
- ✅ Sends email alerts when issues are detected
- ✅ Supports custom date checking
- ✅ Email configuration testing
- ✅ Detailed logging and error reporting

## Installation

1. Install Node.js dependencies:
```bash
npm install
```

2. Set up your email configuration (see Configuration section below)

## Configuration

### Email Settings

You need to configure SMTP settings to send email notifications. Set these environment variables:

```bash
export EMAIL_USER="your-email@gmail.com"
export EMAIL_PASS="your-app-password"
export EMAIL_FROM="backup-monitor@yourdomain.com"  # Optional
export EMAIL_TO="alert-recipient@yourdomain.com"   # Optional
```

### For Gmail Users

1. Enable 2-factor authentication on your Gmail account
2. Generate an App Password:
   - Go to Google Account settings
   - Security → 2-Step Verification → App passwords
   - Generate a password for "Mail"
   - Use this password as `EMAIL_PASS`

### For Other Email Providers

Update the SMTP settings in the `config.email` section of `verifier.js`:

```javascript
email: {
  host: 'your-smtp-server.com',
  port: 587,
  secure: false,
  auth: {
    user: process.env.EMAIL_USER,
    pass: process.env.EMAIL_PASS
  }
}
```

## Usage

### Basic Usage
```bash
# Check today's backup folder
EMAIL_USER=your@email.com EMAIL_PASS=your-password node verifier.js
```

### Check Specific Date
```bash
# Check backup for a specific date
node verifier.js --date=2025-10-25
```

### Test Email Configuration
```bash
# Test if email settings work
node verifier.js --test
```

### Show Help
```bash
node verifier.js --help
```

## Automation

### Cron Job Setup

Add to your crontab to run daily at 9 AM:

```bash
# Edit crontab
crontab -e

# Add this line (adjust paths as needed)
0 9 * * * cd /path/to/backup-verifier && EMAIL_USER=your@email.com EMAIL_PASS=your-password /usr/bin/node verifier.js
```

### Systemd Timer (Alternative)

Create a systemd service and timer for more advanced scheduling:

1. Create service file `/etc/systemd/system/backup-verifier.service`:
```ini
[Unit]
Description=Backup Folder Verifier
After=network.target

[Service]
Type=oneshot
User=bruno
WorkingDirectory=/home/bruno/dev/projects/study-topics/self_hosted/backup/backup-verifier
Environment=EMAIL_USER=your@email.com
Environment=EMAIL_PASS=your-password
ExecStart=/usr/bin/node verifier.js
```

2. Create timer file `/etc/systemd/system/backup-verifier.timer`:
```ini
[Unit]
Description=Run Backup Verifier Daily
Requires=backup-verifier.service

[Timer]
OnCalendar=*-*-* 09:00:00
Persistent=true

[Install]
WantedBy=timers.target
```

3. Enable and start:
```bash
sudo systemctl daemon-reload
sudo systemctl enable backup-verifier.timer
sudo systemctl start backup-verifier.timer
```

## Expected Folder Structure

The script expects backup folders in this format:
```
/media/bruno/KINGSTON/self-hosted-backups/
├── 2025-10-24/
├── 2025-10-25/
├── 2025-10-26/
└── ...
```

## Example Email Notifications

### Missing Folder Alert
```
Subject: 🚨 Backup Alert: Missing backup folder for 2025-10-26

BACKUP VERIFICATION FAILED

Date: 2025-10-26
Folder Path: /media/bruno/KINGSTON/self-hosted-backups/2025-10-26
Issue: Backup folder does not exist
Error: Folder does not exist

Please check your backup system immediately!
```

### Empty Folder Warning
```
Subject: ⚠️ Backup Alert: Empty backup folder for 2025-10-26

BACKUP VERIFICATION WARNING

Date: 2025-10-26
Folder Path: /media/bruno/KINGSTON/self-hosted-backups/2025-10-26
Issue: Backup folder exists but is empty
File Count: 0

The backup folder was created but contains no files.
```

## Troubleshooting

### Common Issues

1. **"Missing required environment variables"**
   - Make sure `EMAIL_USER` and `EMAIL_PASS` are set

2. **"Failed to send email"**
   - Check SMTP settings
   - Verify app password for Gmail
   - Test with `--test` flag

3. **"Permission denied" on folder**
   - Check if the backup drive is mounted
   - Verify read permissions on the backup folder

### Testing

```bash
# Test email configuration
npm run test

# Check specific date
node verifier.js --date=2025-10-25

# Dry run with verbose output
node verifier.js --help
```

## Security Notes

- Never commit email credentials to version control
- Use environment variables or a secure credential store
- Consider using App Passwords instead of main account passwords
- Restrict file permissions on scripts containing credentials

## License

MIT License - Feel free to modify and distribute as needed.
