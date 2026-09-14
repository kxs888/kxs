-- 数据种子目录。
-- S0 不把 argon2id 哈希写进 SQL（盐随机）。演示用户由进程启动时 EnsureBootstrap 创建：
--   BOOTSTRAP_USERNAME / BOOTSTRAP_PASSWORD（仅环境变量）。
-- B7 权限占位（users.permissions 默认）：patient.view / task.create / report.view
-- 禁止在此放入病历全文、患者主档或国密密钥。
SELECT 1;
