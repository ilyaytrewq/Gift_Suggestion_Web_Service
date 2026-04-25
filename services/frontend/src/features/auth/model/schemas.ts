import { z } from 'zod';

export const loginSchema = z.object({
  email: z.string().email('Введите корректный email'),
  password: z.string().min(1, 'Введите пароль'),
});

export const registerSchema = z.object({
  display_name: z
    .string()
    .max(120, 'Имя не должно превышать 120 символов')
    .optional()
    .or(z.literal('')),
  email: z.string().email('Введите корректный email'),
  password: z
    .string()
    .min(8, 'Пароль должен содержать минимум 8 символов')
    .max(72, 'Пароль не должен превышать 72 символа'),
});

export const passwordResetRequestSchema = z.object({
  email: z.string().email('Введите корректный email'),
});

export type LoginSchema = z.infer<typeof loginSchema>;
export type RegisterSchema = z.infer<typeof registerSchema>;
export type PasswordResetRequestSchema = z.infer<
  typeof passwordResetRequestSchema
>;
