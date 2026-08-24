import postgres from 'postgres';

export const sql = postgres(process.env.DATABASE_URL || '', {
  max: 10,
  idle_timeout: 20,
  connect_timeout: 10,
});

export const sqlRead = postgres(process.env.READ_DATABASE_URL || process.env.DATABASE_URL || '', {
  max: 20,
  idle_timeout: 20,
  connect_timeout: 10,
});
