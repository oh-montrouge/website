import { execFileSync, execSync, spawn } from "child_process";
import * as fs from "fs";
import * as http from "http";
import * as path from "path";

const ROOT = path.resolve(__dirname, "..");
const WEBAPP_DIR = path.join(ROOT, "webapp");
const APP_PID_FILE = "/tmp/ohm-e2e-app.pid";

function isServerReady(url: string): Promise<boolean> {
  return new Promise((resolve) => {
    const req = http.get(url, (res) => {
      resolve((res.statusCode ?? 500) < 500);
      res.resume();
    });
    req.on("error", () => resolve(false));
    req.setTimeout(2000, () => {
      req.destroy();
      resolve(false);
    });
  });
}

async function seedAdmin(): Promise<void> {
  // eslint-disable-next-line @typescript-eslint/no-require-imports
  const { Client } = require("pg") as typeof import("pg");
  const adminEmail = process.env.E2E_ADMIN_EMAIL ?? "admin@e2e.local";
  const adminPassword = process.env.E2E_ADMIN_PASSWORD ?? "testpassword";
  const dbUrl = process.env.DATABASE_URL ?? "postgres://ohm:ohm@127.0.0.1:5432/ohm_development?sslmode=disable";

  const hash = execFileSync("go", ["run", ".", adminPassword], {
    cwd: path.join(ROOT, "scripts", "hash-password"),
  })
    .toString()
    .trim();

  const client = new Client({ connectionString: dbUrl });
  await client.connect();
  try {
    const existing = await client.query("SELECT id FROM accounts WHERE email = $1", [adminEmail]);
    if (existing.rows.length > 0) return;

    const instRes = await client.query("SELECT id FROM instruments ORDER BY id LIMIT 1");
    const roleRes = await client.query("SELECT id FROM roles WHERE name = 'admin'");
    const accRes = await client.query(
      "INSERT INTO accounts (email, password_hash, main_instrument_id, status) VALUES ($1, $2, $3, 'active') RETURNING id",
      [adminEmail, hash, instRes.rows[0].id],
    );
    await client.query("INSERT INTO account_roles (account_id, role_id) VALUES ($1, $2)", [
      accRes.rows[0].id,
      roleRes.rows[0].id,
    ]);
  } finally {
    await client.end();
  }
}

export default async function globalSetup(): Promise<void> {
  execSync("docker compose up -d --wait postgres", { cwd: WEBAPP_DIR, stdio: "inherit" });
  execSync("buffalo pop migrate --path ../db/migrations", { cwd: WEBAPP_DIR, stdio: "inherit" });

  await seedAdmin();

  fs.mkdirSync(path.join(ROOT, "coverage", "e2e"), { recursive: true });
  fs.mkdirSync(path.join(ROOT, "bin"), { recursive: true });

  execSync("go build -cover -o ../bin/app-covered ./cmd/app", {
    cwd: WEBAPP_DIR,
    stdio: "inherit",
  });

  let earlyExit: number | null = null;

  const appProc = spawn(path.join(ROOT, "bin", "app-covered"), [], {
    cwd: WEBAPP_DIR,
    env: {
      ...process.env,
      GOCOVERDIR: path.join(ROOT, "coverage", "e2e"),
      GO_ENV: "development",
      SESSION_SECRET: process.env.SESSION_SECRET ?? "dev-secret-change-me",
      DATABASE_URL: process.env.DATABASE_URL ?? "postgres://ohm:ohm@127.0.0.1:5432/ohm_development?sslmode=disable",
    },
    stdio: "ignore",
    detached: true,
  });

  appProc.on("exit", (code) => {
    earlyExit = code ?? 1;
  });
  appProc.unref();
  fs.writeFileSync(APP_PID_FILE, String(appProc.pid));

  try {
    const deadline = Date.now() + 30_000;
    while (Date.now() < deadline) {
      if (earlyExit !== null) throw new Error(`App exited early with code ${earlyExit}`);
      if (await isServerReady("http://localhost:3000/")) return;
      await new Promise((r) => setTimeout(r, 1_000));
    }
    throw new Error("Server not ready at http://localhost:3000/ after 30000ms");
  } catch (err) {
    execSync("docker compose stop", { cwd: WEBAPP_DIR, stdio: "inherit" });
    throw err;
  }
}
