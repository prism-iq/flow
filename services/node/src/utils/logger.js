import pino from 'pino';
import { config } from '../config.js';

export const logger = pino({
  level: config.logLevel,
  transport: process.env.NODE_ENV !== 'production'
    ? { target: 'pino-pretty' }
    : undefined,
  base: {
    service: 'flow-async-io',
  },
  timestamp: pino.stdTimeFunctions.isoTime,
});

export function createChildLogger(context) {
  return logger.child(context);
}
