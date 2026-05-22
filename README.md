# IoT Card Tap System API Documentation

This REST API is built with Golang, Gin Gonic, and GORM. It serves as the backend for an IoT-based door/gate access system utilizing card taps as the access key, with members identified by a generated Unix ID. It supports multiple IoT devices with automatic provisioning and dynamic modes (Active/Register) controllable via a web dashboard (e.g., Next.js).

---

## 🏗️ System Architecture & Workflow

The system is designed to handle multiple IoT readers securely while providing a web dashboard for administration.

### 1. Device Provisioning (First Boot)
When a new IoT device is installed, it doesn't have an identity yet.
1. The device sends a `POST /api/v1/devices/provision` request containing its physical MAC address (`hardware_id`) and a valid provisioning key.
2. The provisioning key can be either `PROVISION_SECRET` or `DEVICE_API_KEY`.
3. The API generates a unique `node_id` (UUID) for this device and registers it in the database with a `pending` state.
4. The device saves this `node_id`.

### 2. Administrator Dashboard (Next.js)
The web admin dashboard interacts with the API using JWT authentication:
- **Admin Login:** Admin logs in and receives a JWT Access Token.
- **Device Activation:** The admin sees the `pending` device on the dashboard and changes its mode to `active`.
- **Member Management:** The admin can add, edit, or delete registered card members.

### 3. Tap to Unlock (Active Mode)
This is the normal day-to-day operation.
1. A user taps their card on the IoT device.
2. The device sends a `POST /api/v1/card/tap` request with the member's `unix_id`. It authenticates using headers: `X-Node-ID` and `X-API-Key`.
3. The API checks if the Unix ID exists in the `members` table and is active.
4. If valid, the API responds with `action: "granted"`. The device unlocks the door. If invalid, it responds with `action: "denied"`.
5. Every tap is recorded in the `access_logs` table.

### 4. Tap to Register (Register Mode)
Instead of typing IDs manually on devices, admins register members in the dashboard first to receive a Unix ID.
1. Admin creates a member in the dashboard using name and phone number, then receives a generated `unix_id`.
2. Admin changes a specific device's mode to `register` via the dashboard.
3. The device periodically polls `GET /api/v1/devices/me/status` and realizes its mode changed.
4. A user taps a card on the device. The device sends the tap request with the `unix_id`.
5. The API validates the Unix ID (and optional name/phone match) and activates the member if valid.

---

## 🔐 Authentication & API Keys

The system uses different authentication methods depending on the client.

### For Web Dashboard (Users)
- **JWT (JSON Web Tokens):** Uses short-lived Access Tokens (15 mins) and long-lived Refresh Tokens (7 days). Sent via the `Authorization: Bearer <token>` header.

### For IoT Devices
IoT devices authenticate using a 3-tier API Key system for flexibility and security. Devices must always send two headers on every request:
- `X-Node-ID`: The unique ID received during provisioning.
- `X-API-Key`: The authentication key.

**The 3-Tier API Key Verification:**
1. **Master Key (`DEVICE_MASTER_KEY`):** If provided in `.env` and matches the request, it bypasses all checks. *(For development only)*
2. **Global Key (`DEVICE_API_KEY`):** A single master key stored in `.env`. All IoT devices can use this exact same key. This is highly recommended as it simplifies IoT firmware configuration.
3. **Per-Device Key:** A unique key generated per device in the database during provisioning. Used if no global key is set.

---

## ⚙️ Environment Configuration (`.env`)

Here is the explanation of all variables inside the `.env` file.

| Variable | Description | Example |
|---|---|---|
| `PORT` | The port the Go API runs on. | `8080` |
| `APP_ENV` | Environment mode (`development` or `production`). | `development` |
| `DB_HOST` | MySQL database host. | `127.0.0.1` |
| `DB_PORT` | MySQL database port. | `3306` |
| `DB_USER` | MySQL database user. | `root` |
| `DB_PASSWORD` | MySQL database password. | *(empty for Laragon default)* |
| `DB_NAME` | MySQL database name. Automatically migrated on boot. | `iot_ktp` |
| `JWT_SECRET` | Secret key used to sign and verify JWT tokens for the web dashboard. | `e39d4...6cd0` |
| `JWT_ACCESS_DURATION_MIN` | JWT Access Token lifespan in minutes. | `15` |
| `JWT_REFRESH_DURATION_DAYS` | JWT Refresh Token lifespan in days. | `7` |
| `PROVISION_SECRET` | Optional dedicated provisioning secret. If you want a single-key setup, the firmware may also use `DEVICE_API_KEY` for provisioning. | `b3bb2...bdbe` |
| `DEVICE_API_KEY` | The Global API Key for IoT devices. Devices use this as the `X-API-Key` header when sending Tap events. | `8e355...303a5` |
| `DEVICE_MASTER_KEY` | Development master key. Bypasses device validation. **Must be empty in production.** | *(empty)* |
| `REGISTER_MODE_TIMEOUT_MIN` | Minutes before a device in `register` mode automatically reverts to `active`. `0` means manual only. | `0` |
| `ALLOWED_ORIGINS` | CORS origins allowed to access the API. Separate with commas. | `http://localhost:3000` |

---

## 🚀 Getting Started (Development)

1. Ensure MySQL is running and the credentials match your `.env`.
2. Install dependencies: `go mod tidy`
3. Run the server: `go run ./cmd/main.go`
4. The server will auto-migrate the database tables upon starting and seed the first admin user using `ADMIN_EMAIL` and `ADMIN_PASSWORD` from `.env`.

---

## 🌍 Deployment Guide

### Deploying on Ubuntu (Linux)

**1. Install Go & MySQL**
```bash
sudo apt update
sudo apt install golang mysql-server -y
```

**2. Setup Database**
```bash
sudo mysql
mysql> CREATE DATABASE iot_ktp;
mysql> EXIT;
```

**3. Build the Application**
```bash
# Clone the repository and cd into it
go mod tidy
go build -o iot-api ./cmd/main.go
```

**4. Setup Environment File**
```bash
cp .env.example .env
nano .env # Edit the database credentials and passwords
```

**5. Setup Systemd Service (Run in background)**
Create a service file: `sudo nano /etc/systemd/system/iot-api.service`
```ini
[Unit]
Description=IoT Card API Service
After=network.target mysql.service

[Service]
User=root
WorkingDirectory=/path/to/your/app
ExecStart=/path/to/your/app/iot-api
Restart=always

[Install]
WantedBy=multi-user.target
```
Start and enable the service:
```bash
sudo systemctl daemon-reload
sudo systemctl enable iot-api
sudo systemctl start iot-api
sudo systemctl status iot-api
```

### Deploying on Windows (Using Laragon / PM2 / NSSM)

**1. Build the Executable**
Open PowerShell or Command Prompt in the project directory:
```powershell
go build -o iot-api.exe ./cmd/main.go
```

**2. Setup Database**
Ensure Laragon is running with MySQL started. Verify that the `iot_ktp` database exists.

**3. Setup Environment File**
Copy `.env.example` to `.env` and configure your Laragon MySQL credentials (default user is `root`, password empty).

**4. Run as a Background Service**
You can use **NSSM (Non-Sucking Service Manager)** to install it as a Windows Service:
1. Download NSSM and extract it.
2. Open PowerShell as Administrator and run:
   ```powershell
   nssm install IoT-Card-API "C:\path\to\your\app\iot-api.exe"
   ```
3. Set the AppDirectory to the folder containing your `.env` file (`C:\path\to\your\app`).
4. Start the service:
   ```powershell
   nssm start IoT-Card-API
   ```

*Alternatively, you can run it using PM2 if you have Node.js installed:*
```powershell
npm install -g pm2
pm2 start iot-api.exe --name "iot-card-api"
pm2 save
```
