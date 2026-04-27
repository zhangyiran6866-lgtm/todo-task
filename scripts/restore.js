const { execSync } = require('child_process');
const path = require('path');
const fs = require('fs');
const yaml = require('js-yaml');
const readline = require('readline');

/**
 * MongoDB 自动恢复脚本
 * 职责：
 * 1. 扫描 backups 目录供用户选择
 * 2. 获取备份文件中的数据库名称
 * 3. 执行 mongorestore（带 --drop 安全覆盖）
 */

const CONFIG_PATH = path.join(__dirname, '../packages/backend/configs/config.yaml');
const BACKUP_ROOT = path.join(__dirname, '../backups');

const rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout
});

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

async function runRestore() {
  if (!fs.existsSync(BACKUP_ROOT)) {
    console.error('❌ 未找到 backups 目录');
    process.exit(1);
  }

  const folders = fs.readdirSync(BACKUP_ROOT).filter(f => f.startsWith('backup-'));
  
  if (folders.length === 0) {
    console.error('❌ backups 目录下没有可用的备份文件');
    process.exit(1);
  }

  console.log('\n--- 可用备份列表 ---');
  folders.sort().reverse().forEach((folder, index) => {
    console.log(`[${index + 1}] ${folder}`);
  });
  console.log('-------------------\n');

  rl.question('请输入要恢复的备份编号: ', (answer) => {
    const index = parseInt(answer) - 1;
    if (isNaN(index) || index < 0 || index >= folders.length) {
      console.error('❌ 无效的编号');
      rl.close();
      return;
    }

    const selectedFolder = folders[index];
    const backupPath = path.join(BACKUP_ROOT, selectedFolder);

    // 探测备份目录下的数据库文件夹
    const dbFolders = fs.readdirSync(backupPath).filter(f => {
      return fs.statSync(path.join(backupPath, f)).isDirectory();
    });

    if (dbFolders.length === 0) {
      console.error('❌ 选中的备份目录中未找到数据库数据');
      rl.close();
      return;
    }

    const dbNameInBackup = dbFolders[0]; // 默认取第一个
    const { uri } = getMongoConfig();

    console.log(`\n⚠️  危险操作：即将恢复备份 [${selectedFolder}]`);
    console.log(`⚠️  这会【删除并覆盖】当前数据库 [${dbNameInBackup}] 中的所有数据！`);
    
    rl.question('确认继续？(y/n): ', (confirm) => {
      if (confirm.toLowerCase() !== 'y') {
        console.log('🚫 操作已取消');
        rl.close();
        return;
      }

      console.log('🚀 正在执行恢复...');

      try {
        // 执行 mongorestore
        // --drop 在恢复前删除集合
        // --uri 使用连接串
        const command = `mongorestore --uri="${uri}" --drop "${backupPath}"`;
        
        execSync(command, { stdio: 'inherit' });
        
        console.log('\n✅ 数据库恢复成功！');
      } catch (error) {
        console.error('\n❌ 恢复失败:', error.message);
      } finally {
        rl.close();
      }
    });
  });
}

runRestore();
