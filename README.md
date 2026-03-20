# Attendance Tracking System

> Smart student attendance system using RFID / mobile devices with real-time processing and reporting

---

## About

This project is a **student attendance tracking system** designed to automate the process of recording and analyzing attendance using unique identifiers (RFID cards, mobile devices, etc.).

The system focuses on:
- fast data collection
- reliable storage
- flexible reporting
- scalability and integration with external devices

---

## How It Works

The system is divided into two main parts:

### 1. Data Collection (Client Side)
- students scan their ID (RFID / device)
- system records:
    - unique identifier
    - timestamp
    - location (classroom)
- data is stored locally and sent to the server

---

### 2. Backend Processing (Server Side)
- receives attendance data
- stores it in database
- processes and validates records
- generates reports

> Client only collects data — all logic is handled on the server.

---

## Project Structure

```bash
attendance-system/
│
├── cmd/                    # Application entry points
│   └── server/
│
├── internal/
│   ├── domain/             # Core entities (Student, Attendance)
│   ├── service/            # Business logic
│   ├── storage/            # Database layer
│   │   └── postgres/
│   └── transport/          # HTTP handlers / API
│
├── pkg/                    # Shared packages
├── configs/                # Config files
├── migrations/             # DB migrations
├── docs/                   # Documentation
└── README.md