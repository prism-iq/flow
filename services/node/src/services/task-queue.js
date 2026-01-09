import EventEmitter from 'eventemitter3';
import { logger } from '../utils/logger.js';

export class TaskQueue extends EventEmitter {
  constructor(config) {
    super();
    this.config = config;
    this.tasks = new Map();
    this.pending = [];
    this.processing = new Set();
    this.handlers = new Map();
  }

  async enqueue(type, payload, priority = 0) {
    const taskId = `task_${Date.now()}_${Math.random().toString(36).slice(2)}`;

    const task = {
      id: taskId,
      type,
      payload,
      priority,
      status: 'pending',
      createdAt: Date.now(),
      attempts: 0,
    };

    this.tasks.set(taskId, task);
    this.pending.push(taskId);
    this.pending.sort((a, b) => {
      const taskA = this.tasks.get(a);
      const taskB = this.tasks.get(b);
      return taskB.priority - taskA.priority;
    });

    this.emit('enqueued', task);
    this.processNext();

    return taskId;
  }

  async processNext() {
    if (this.processing.size >= this.config.maxConcurrency) {
      return;
    }

    const taskId = this.pending.shift();
    if (!taskId) return;

    const task = this.tasks.get(taskId);
    if (!task) return;

    this.processing.add(taskId);
    task.status = 'processing';
    task.startedAt = Date.now();

    try {
      const handler = this.handlers.get(task.type);
      if (!handler) {
        throw new Error(`No handler for task type: ${task.type}`);
      }

      const result = await handler(task.payload);

      task.status = 'completed';
      task.result = result;
      task.completedAt = Date.now();

      this.emit('completed', task);

    } catch (error) {
      task.attempts++;
      task.error = error.message;

      if (task.attempts < this.config.retryAttempts) {
        task.status = 'pending';
        setTimeout(() => {
          this.pending.push(taskId);
          this.processNext();
        }, this.config.retryDelay * task.attempts);
      } else {
        task.status = 'failed';
        this.emit('failed', task);
      }

      logger.error({ taskId, error: error.message }, 'Task failed');
    }

    this.processing.delete(taskId);
    this.processNext();
  }

  registerHandler(type, handler) {
    this.handlers.set(type, handler);
  }

  async getStatus(taskId) {
    return this.tasks.get(taskId);
  }

  async cancel(taskId) {
    const task = this.tasks.get(taskId);
    if (!task) return false;

    if (task.status === 'pending') {
      task.status = 'cancelled';
      this.pending = this.pending.filter(id => id !== taskId);
      return true;
    }

    return false;
  }

  getStats() {
    const stats = {
      total: this.tasks.size,
      pending: this.pending.length,
      processing: this.processing.size,
      completed: 0,
      failed: 0,
    };

    for (const task of this.tasks.values()) {
      if (task.status === 'completed') stats.completed++;
      if (task.status === 'failed') stats.failed++;
    }

    return stats;
  }

  cleanup(maxAge = 3600000) {
    const now = Date.now();
    for (const [taskId, task] of this.tasks) {
      if (task.status === 'completed' || task.status === 'failed') {
        if (now - task.createdAt > maxAge) {
          this.tasks.delete(taskId);
        }
      }
    }
  }
}
