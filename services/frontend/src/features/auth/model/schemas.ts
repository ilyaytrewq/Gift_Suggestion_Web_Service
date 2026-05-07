import { z } from 'zod';

export const loginSchema = z.object({
  email: z.string().email('Введите корректный email'),
  password: z.string().min(1, 'Введите пароль'),
});

const passwordSchema = z
  .string()
  .min(8, 'Пароль должен содержать минимум 8 символов')
  .max(72, 'Пароль не должен превышать 72 символа')
  .regex(/[a-z]/, 'Пароль должен содержать строчную букву')
  .regex(/[A-Z]/, 'Пароль должен содержать заглавную букву')
  .regex(/[0-9]/, 'Пароль должен содержать цифру')
  .regex(/[^a-zA-Z0-9]/, 'Пароль должен содержать специальный символ')
  .regex(/^\S*$/, 'Пароль не должен содержать пробелы');

export const registerSchema = z.object({
  display_name: z
    .string()
    .max(120, 'Имя не должно превышать 120 символов')
    .optional()
    .or(z.literal('')),
  email: z.string().email('Введите корректный email'),
  password: passwordSchema,
});

export const passwordResetRequestSchema = z.object({
  email: z.string().email('Введите корректный email'),
});

export const passwordResetConfirmSchema = z.object({
  new_password: passwordSchema,
});

export type LoginSchema = z.infer<typeof loginSchema>;
export type RegisterSchema = z.infer<typeof registerSchema>;
export type PasswordResetRequestSchema = z.infer<
  typeof passwordResetRequestSchema
>;
export type PasswordResetConfirmSchema = z.infer<typeof passwordResetConfirmSchema>;
