const { execSync } = require('child_process');
const path = require('path');
const fs = require('fs');
const yaml = require('js-yaml');

/**
 * MongoDB 自动备份脚本
 * 职责：
 * 1. 读取配置文件获取连接串
 * 2. 执行 mongodump 生成备份
 * 3. 自动清理过期备份（保留 7 天）
 */

const CONFIG_PATH = path.join(__dirname, '../packages/backend/configs/config.yaml');
const BACKUP_ROOT = path.join(__dirname, '../backups');
const RETENTION_DAYS = 7;

function getMongoConfig() {
  try {
    const fileContents = fs.readFileSync(CONFIG_PATH, 'utf8');
    const data = yaml.load(fileContents);
    return {
      uri: data.mongodb.uri,
      database: data.mongodb.database,
    };
  } catch (e) {
    console.error('❌ 无法读取配置文件:', e.message);
    process.exit(1);
  }
}

function cleanOldBackups() {
  console.log('🧹 正在检查并清理过期备份...');
  const now = Date.now();
  const folders = fs.readdirSync(BACKUP_ROOT);

  folders.forEach((folder) => {
    const folderPath = path.join(BACKUP_ROOT, folder);
    const stats = fs.statSync(folderPath);
    
    // 计算文件夹年龄
    const ageInDays = (now - stats.mtimeMs) / (1000 * 60 * 60 * 24);
    
    if (ageInDays > RETENTION_DAYS) {
      console.log(`🗑️ 删除过期备份: ${folder} (创建于 ${ageInDays.toFixed(1)} 天前)`);
      fs.rmSync(folderPath, { recursive: true, force: true });
    }
  });
}

function runBackup() {
  const { uri, database } = getMongoConfig();
  const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
  const backupPath = path.join(BACKUP_ROOT, `backup-${timestamp}`);

  console.log(`🚀 开始备份数据库 [${database}] 到: ${backupPath}`);

  try {
    // 确保备份根目录存在
    if (!fs.existsSync(BACKUP_ROOT)) {
      fs.mkdirSync(BACKUP_ROOT, { recursive: true });
    }

    // 执行 mongodump
    // --uri 使用连接字符串
    // --out 指定输出目录
    const command = `mongodump --uri="${uri}" --out="${backupPath}"`;
    
    execSync(command, { stdio: 'inherit' });
    
    console.log('✅ 备份完成！');
    
    // 执行清理
    cleanOldBackups();
  } catch (error) {
    console.error('❌ 备份失败:', error.message);
    process.exit(1);
  }
}

// 执行脚本
runBackup();
