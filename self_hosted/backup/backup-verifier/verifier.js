#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const nodemailer = require('nodemailer');

// Configuration
const config = {
  // Email configuration
  email: {
    service: 'gmail',
    host: 'smtp.gmail.com', // Change this to your SMTP server
    port: 465,
    secure: true, // true for 465, false for other ports
    auth: {
      user: process.env.EMAIL_USER || 'your-email@gmail.com', // Set via environment variable
      pass: process.env.EMAIL_PASS || 'your-app-password'     // Set via environment variable
    }
  },
  
  // Notification settings
  notification: {
    from: process.env.EMAIL_FROM || 'backup-monitor@yourdomain.com',
    to: process.env.EMAIL_TO || 'your-email@gmail.com', // Recipient email
    subject: 'Backup Verification Alert'
  },
  
  // Backup folder path template
  backupBasePath: '/media/bruno/KINGSTON/self-hosted-backups'
};

/**
 * Get current date in YYYY-MM-DD format
 */
function getCurrentDateString() {
  const now = new Date();
  const year = now.getFullYear();
  const month = String(now.getMonth() + 1).padStart(2, '0');
  const day = String(now.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
}

/**
 * Check if a directory exists and is not empty
 */
function checkBackupFolder(folderPath) {
  try {
    // Check if folder exists
    if (!fs.existsSync(folderPath)) {
      return {
        exists: false,
        isEmpty: true,
        error: 'Folder does not exist'
      };
    }

    // Check if it's actually a directory
    const stats = fs.statSync(folderPath);
    if (!stats.isDirectory()) {
      return {
        exists: false,
        isEmpty: true,
        error: 'Path exists but is not a directory'
      };
    }

    // Check if directory is empty
    const files = fs.readdirSync(folderPath);
    const isEmpty = files.length === 0;

    return {
      exists: true,
      isEmpty: isEmpty,
      fileCount: files.length,
      files: files
    };
  } catch (error) {
    return {
      exists: false,
      isEmpty: true,
      error: error.message
    };
  }
}

/**
 * Send email notification
 */
async function sendEmailNotification(subject, message) {
  try {
    // Create transporter
    const transporter = nodemailer.createTransport(config.email);

    // Email options
    const mailOptions = {
      from: config.notification.from,
      to: config.notification.to,
      subject: subject,
      text: message,
      html: `<pre>${message}</pre>`
    };

    // Send email
    const info = await transporter.sendMail(mailOptions);
    console.log('Email sent successfully:', info.messageId);
    return true;
  } catch (error) {
    console.error('Failed to send email:', error.message);
    return false;
  }
}

/**
 * Main verification function
 */
async function verifyBackup() {
  const dateString = getCurrentDateString();
  const backupFolderPath = path.join(config.backupBasePath, dateString);
  
  console.log(`Checking backup folder: ${backupFolderPath}`);
  
  const result = checkBackupFolder(backupFolderPath);
  
  let shouldNotify = false;
  let notificationMessage = '';
  let subject = '';

  if (!result.exists) {
    shouldNotify = true;
    subject = `🚨 Backup Alert: Missing backup folder for ${dateString}`;
    notificationMessage = `BACKUP VERIFICATION FAILED

Date: ${dateString}
Folder Path: ${backupFolderPath}
Issue: Backup folder does not exist
Error: ${result.error || 'Unknown error'}

Please check your backup system immediately!

---
Backup Verification System
${new Date().toISOString()}`;
    
    console.log('❌ Backup folder does not exist');
  } else if (result.isEmpty) {
    shouldNotify = true;
    subject = `⚠️ Backup Alert: Empty backup folder for ${dateString}`;
    notificationMessage = `BACKUP VERIFICATION WARNING

Date: ${dateString}
Folder Path: ${backupFolderPath}
Issue: Backup folder exists but is empty
File Count: ${result.fileCount}

The backup folder was created but contains no files. Please verify your backup process.

---
Backup Verification System
${new Date().toISOString()}`;
    
    console.log('⚠️ Backup folder exists but is empty');
  } else {
    console.log(`✅ Backup folder exists and contains ${result.fileCount} files`);
    console.log('Files found:', result.files.join(', '));
  }

  if (shouldNotify) {
    console.log('Sending notification email...');
    const emailSent = await sendEmailNotification(subject, notificationMessage);
    
    if (emailSent) {
      console.log('✅ Email notification sent successfully');
    } else {
      console.log('❌ Failed to send email notification');
      process.exit(1);
    }
  } else {
    console.log('✅ Backup verification passed - no notification needed');
  }
}

/**
 * Display help information
 */
function showHelp() {
  console.log(`
Backup Folder Verifier

This script checks if the backup folder for the current date exists and contains files.
If the folder is missing or empty, it sends an email notification.

Usage:
  node verifier.js [options]

Options:
  --help, -h     Show this help message
  --test         Test email configuration
  --date=DATE    Check specific date (YYYY-MM-DD format)

Environment Variables:
  EMAIL_USER     SMTP username (required)
  EMAIL_PASS     SMTP password/app password (required)
  EMAIL_FROM     Sender email address
  EMAIL_TO       Recipient email address

Example:
  EMAIL_USER=your-email@gmail.com EMAIL_PASS=your-app-password node verifier.js
  node verifier.js --date=2025-10-25
  node verifier.js --test
`);
}

/**
 * Test email configuration
 */
async function testEmailConfig() {
  console.log('Testing email configuration...');
  
  const testMessage = `This is a test email from the Backup Verifier script.

Configuration:
- SMTP Host: ${config.email.host}
- SMTP Port: ${config.email.port}
- From: ${config.notification.from}
- To: ${config.notification.to}

If you receive this email, your configuration is working correctly!

---
Test sent at: ${new Date().toISOString()}`;

  const success = await sendEmailNotification('📧 Backup Verifier - Test Email', testMessage);
  
  if (success) {
    console.log('✅ Email test successful!');
  } else {
    console.log('❌ Email test failed!');
    process.exit(1);
  }
}

// Main execution
async function main() {
  //console.log(config)
  const args = process.argv.slice(2);
  
  // Check for help flag
  if (args.includes('--help') || args.includes('-h')) {
    showHelp();
    return;
  }
  
  // Check for test flag
  if (args.includes('--test')) {
    await testEmailConfig();
    return;
  }
  
  // Check for custom date
  const dateArg = args.find(arg => arg.startsWith('--date='));
  if (dateArg) {
    const customDate = dateArg.split('=')[1];
    if (!/^\d{4}-\d{2}-\d{2}$/.test(customDate)) {
      console.error('❌ Invalid date format. Please use YYYY-MM-DD format.');
      process.exit(1);
    }
    
    // Override getCurrentDateString function for custom date
    const originalGetCurrentDateString = getCurrentDateString;
    getCurrentDateString = () => customDate;
    console.log(`Using custom date: ${customDate}`);
  }
  
  // Verify required environment variables
  if (!process.env.EMAIL_USER || !process.env.EMAIL_PASS) {
    console.error('❌ Missing required environment variables:');
    console.error('   EMAIL_USER and EMAIL_PASS must be set');
    console.error('   Run with --help for more information');
    process.exit(1);
  }
  
  try {
    await verifyBackup();
  } catch (error) {
    console.error('❌ Script execution failed:', error.message);
    process.exit(1);
  }
}

// Run the script
if (require.main === module) {
  main().catch(error => {
    console.error('❌ Unhandled error:', error);
    process.exit(1);
  });
}
