import { logger } from '../utils/logger.js';

export function errorHandler(error, request, reply) {
  logger.error({
    error: error.message,
    stack: error.stack,
    method: request.method,
    url: request.url,
  }, 'Request error');

  if (error.validation) {
    return reply.code(400).send({
      error: 'Validation Error',
      message: error.message,
      details: error.validation,
    });
  }

  if (error.statusCode) {
    return reply.code(error.statusCode).send({
      error: error.name || 'Error',
      message: error.message,
    });
  }

  return reply.code(500).send({
    error: 'Internal Server Error',
    message: process.env.NODE_ENV === 'production'
      ? 'An unexpected error occurred'
      : error.message,
  });
}
