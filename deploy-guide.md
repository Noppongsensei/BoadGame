# คู่มือการ Deploy Boadgame บน VPS

คู่มือนี้อธิบายขั้นตอนการติดตั้งและเรียกใช้งาน Boadgame บน VPS โดยใช้ Docker และ Caddy

## ข้อกำหนดระบบ

- VPS ที่มี Docker และ Docker Compose ติดตั้งแล้ว
- OS: Ubuntu 20.04+ หรือ Debian 11+
- CPU: อย่างน้อย 2 cores
- RAM: อย่างน้อย 4 GB
- Disk: อย่างน้อย 20 GB
- โดเมนเนมที่ชี้มาที่ IP ของเซิร์ฟเวอร์

## ขั้นตอนการ Deploy

### 1. Clone Repository

```bash
# สร้างโฟลเดอร์สำหรับแอปพลิเคชัน
mkdir -p /opt/boadgame
cd /opt/boadgame

# Clone repository
git clone https://github.com/yourusername/boadgame.git .
```

### 2. สร้างไฟล์ .env

สร้างไฟล์ `.env` ในโฟลเดอร์หลักของโปรเจค:

```bash
cat > .env << EOF
POSTGRES_USER=postgres
POSTGRES_PASSWORD=your_secure_password_here
POSTGRES_DB=avalon
JWT_SECRET=your_secure_jwt_secret_here
NEXT_PUBLIC_API_URL=
CORS_ORIGINS=https://your-domain.com
EOF
```

โปรดแน่ใจว่าได้เปลี่ยน `your_secure_password_here`, `your_secure_jwt_secret_here` และ `your-domain.com` เป็นค่าที่เหมาะสม

### 3. แก้ไขไฟล์ Caddyfile

แก้ไข domain ในไฟล์ `Caddyfile`:

```bash
# แก้ไขโดเมนในไฟล์ Caddyfile
sed -i 's/avalon-thai.com/your-domain.com/g' Caddyfile
```

แทนที่ `your-domain.com` ด้วยโดเมนจริงที่จะใช้งาน

### 4. Build และ Start Docker Containers

```bash
# สร้าง containers และรันในโหมด detached
docker compose -f docker-compose.prod.yml --env-file .env up -d --build
```

### 5. ตรวจสอบสถานะ Containers

```bash
# ตรวจสอบว่า containers ทำงานอยู่หรือไม่
docker compose -f docker-compose.prod.yml ps
```

### 6. ตั้งค่า DNS และ Cloudflare (ถ้าใช้)

1. เพิ่ม A record ที่ชี้ไปยัง IP ของ VPS
2. ตั้งค่า Cloudflare SSL/TLS:
   - Full mode (ไม่ใช่ Flexible หรือ Strict)
   - Always Use HTTPS: เปิดใช้งาน
   - Minimum TLS Version: 1.2

### 7. การเข้าถึงแอปพลิเคชัน

- ผู้ใช้สามารถเข้าถึงแอปพลิเคชันได้ที่ `https://your-domain.com`
- WebSocket จะทำงานที่ `wss://your-domain.com/ws`
- API endpoints จะทำงานที่ `https://your-domain.com/api/*`

## การบำรุงรักษาและการแก้ไขปัญหา

### การอัปเดตแอปพลิเคชัน

```bash
# Pull การเปลี่ยนแปลงล่าสุดจาก repository
cd /opt/boadgame
git pull

# รีบิลด์และรีสตาร์ท containers
docker compose -f docker-compose.prod.yml --env-file .env up -d --build
```

### การดู Logs

```bash
# ดู logs ของทุก containers
docker compose -f docker-compose.prod.yml logs

# ดู logs แบบต่อเนื่อง
docker compose -f docker-compose.prod.yml logs -f

# ดู logs ของ container เฉพาะ
docker compose -f docker-compose.prod.yml logs backend
docker compose -f docker-compose.prod.yml logs frontend
docker compose -f docker-compose.prod.yml logs caddy
```

### การแก้ไขปัญหา

#### หากแอปพลิเคชันไม่ทำงาน

1. ตรวจสอบว่า containers ทำงานอยู่หรือไม่:
```bash
docker compose -f docker-compose.prod.yml ps
```

2. ดู logs เพื่อหาข้อผิดพลาด:
```bash
docker compose -f docker-compose.prod.yml logs
```

3. รีสตาร์ท containers:
```bash
docker compose -f docker-compose.prod.yml restart
```

#### หากการเชื่อมต่อ WebSocket มีปัญหา

1. ตรวจสอบ logs ของ backend:
```bash
docker compose -f docker-compose.prod.yml logs backend
```

2. ตรวจสอบว่า port 8080 ภายใน container เปิดอยู่:
```bash
docker compose -f docker-compose.prod.yml exec backend netstat -tuln
```

## การ Rollback

หากมีปัญหาหลังการอัปเดต คุณสามารถ rollback กลับไปยังเวอร์ชันก่อนหน้าได้:

```bash
# ดู git commit history
cd /opt/boadgame
git log --oneline

# ย้อนกลับไปยัง commit เวอร์ชันก่อนหน้า
git checkout <commit-hash>

# รีบิลด์และรีสตาร์ท containers
docker compose -f docker-compose.prod.yml --env-file .env up -d --build
```

## การสำรองข้อมูล

### การสำรองข้อมูล PostgreSQL

```bash
# สำรองข้อมูล PostgreSQL
docker compose -f docker-compose.prod.yml exec -T postgres pg_dump -U postgres avalon > backup_$(date +%Y%m%d_%H%M%S).sql
```

### การกู้คืนข้อมูล PostgreSQL

```bash
# กู้คืนข้อมูล PostgreSQL
cat backup_file.sql | docker compose -f docker-compose.prod.yml exec -T postgres psql -U postgres avalon
```

## การดูแลระบบในระยะยาว

- ตรวจสอบพื้นที่ดิสก์เป็นประจำ: `df -h`
- ทำการสำรองข้อมูลเป็นประจำ
- อัปเดต OS และ Docker เป็นประจำ
- ตั้งค่า monitoring system เพื่อแจ้งเตือนเมื่อมีปัญหา

---

หากมีข้อสงสัยหรือต้องการความช่วยเหลือเพิ่มเติม โปรดติดต่อทีมพัฒนา
