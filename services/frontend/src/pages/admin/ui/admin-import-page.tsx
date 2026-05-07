import { useRef, useState } from 'react';
import { useQuery } from '@tanstack/react-query';

import type { ImportJob, CreateImportJobResponse } from '../../../shared/api/contracts';
import { requestJson } from '../../../shared/api/http';
import { useAuth } from '../../../shared/auth/use-auth';
import { Button } from '../../../shared/ui/button/button';
import { ErrorBanner } from '../../../shared/ui/feedback/error-banner';
import { Notice } from '../../../shared/ui/feedback/notice';
import { Container } from '../../../shared/ui/layout/container';

// ─── API helpers ─────────────────────────────────────────────────────────────

async function createImportJob(file: File, source?: string): Promise<CreateImportJobResponse> {
  const form = new FormData();
  form.append('file', file);
  if (source) form.append('source', source);

  return requestJson<CreateImportJobResponse>('/api/v1/admin/import-jobs', {
    method: 'POST',
    body: form,
    auth: true,
  });
}

function getImportJob(jobId: string) {
  return requestJson<{ data: { job: ImportJob } }>(`/api/v1/admin/import-jobs/${jobId}`, {
    auth: true,
  });
}

function getImportJobErrors(jobId: string) {
  return requestJson<{ data: { errors: Array<{ row: number; message: string }> } }>(
    `/api/v1/admin/import-jobs/${jobId}/errors`,
    { auth: true },
  );
}

// ─── Status display ───────────────────────────────────────────────────────────

const STATUS_LABELS: Record<string, string> = {
  pending: 'Ожидает',
  running: 'Выполняется',
  completed: 'Завершено',
  completed_with_errors: 'Завершено с ошибками',
  failed: 'Ошибка',
};

const ACTIVE_STATUSES = new Set(['pending', 'running']);

// ─── Component ───────────────────────────────────────────────────────────────

export function AdminImportPage(): JSX.Element {
  const auth = useAuth();
  const fileRef = useRef<HTMLInputElement>(null);
  const [source, setSource] = useState('');
  const [uploading, setUploading] = useState(false);
  const [uploadError, setUploadError] = useState<Error | null>(null);
  const [jobId, setJobId] = useState<string | null>(null);

  const jobQuery = useQuery({
    enabled: Boolean(jobId),
    queryKey: ['import-job', jobId],
    queryFn: () => getImportJob(jobId!),
    refetchInterval: (query) => {
      const status = query.state.data?.data.job.status;
      return status && ACTIVE_STATUSES.has(status) ? 2000 : false;
    },
  });

  const errorsQuery = useQuery({
    enabled: Boolean(jobId) && jobQuery.data?.data.job.status === 'completed_with_errors',
    queryKey: ['import-job-errors', jobId],
    queryFn: () => getImportJobErrors(jobId!),
  });

  const job = jobQuery.data?.data.job;

  async function handleUpload() {
    const file = fileRef.current?.files?.[0];
    if (!file) return;
    setUploadError(null);
    setUploading(true);
    try {
      const res = await createImportJob(file, source || undefined);
      setJobId(res.data.job.id);
    } catch (err) {
      setUploadError(err instanceof Error ? err : new Error(String(err)));
    } finally {
      setUploading(false);
    }
  }

  if (!auth.accessToken || auth.user?.role !== 'admin') {
    return (
      <Container className="auth-page">
        <p>Для доступа к этому разделу необходимо войти как администратор.</p>
      </Container>
    );
  }

  return (
    <Container className="page-stack import-page">
      <section className="section-heading">
        <p className="eyebrow">Администрирование</p>
        <h1>Импорт каталога</h1>
        <p className="page-copy">
          Загрузите файл CSV, JSON или XLSX для пополнения каталога подарков.
        </p>
      </section>

      <div className="import-form">
        <div className="field">
          <label className="field__label" htmlFor="import-file">
            Файл каталога <span style={{ color: 'var(--color-error)' }}>*</span>
          </label>
          <input
            accept=".csv,.json,.xlsx"
            className="input"
            id="import-file"
            ref={fileRef}
            type="file"
          />
          <p className="field__hint">Поддерживаются форматы: CSV, JSON, XLSX</p>
        </div>

        <div className="field">
          <label className="field__label" htmlFor="import-source">
            Источник (необязательно)
          </label>
          <input
            className="input"
            id="import-source"
            placeholder="Например: wildberries_2026_04"
            type="text"
            value={source}
            onChange={(e) => setSource(e.target.value)}
          />
        </div>

        {uploadError && <ErrorBanner error={uploadError} title="Ошибка загрузки" />}

        <Button
          disabled={uploading}
          type="button"
          onClick={() => void handleUpload()}
        >
          {uploading ? 'Загружаем…' : 'Загрузить файл'}
        </Button>
      </div>

      {job && (
        <div className="import-job">
          <div className="import-job__header">
            <strong>Задача</strong>{' '}
            <span className="import-job__filename">{job.source_filename}</span>
          </div>

          <div>
            <span
              className={`import-job__status import-job__status--${job.status}`}
            >
              {STATUS_LABELS[job.status] ?? job.status}
            </span>
            {ACTIVE_STATUSES.has(job.status) && (
              <span style={{ marginLeft: '0.5rem', fontSize: '0.8rem', color: 'var(--color-muted)' }}>
                (обновляется автоматически)
              </span>
            )}
          </div>

          {job.failure_message && (
            <Notice tone="error">{job.failure_message}</Notice>
          )}

          {job.summary && (
            <div className="import-job__stats">
              {([
                ['Всего строк', job.summary.total_rows],
                ['Обработано', job.summary.processed_rows],
                ['Импортировано', job.summary.imported_rows],
                ['Обновлено', job.summary.updated_rows],
                ['Пропущено', job.summary.skipped_rows],
                ['Ошибок', job.summary.error_rows],
              ] as [string, number][]).map(([label, value]) => (
                <div className="import-job__stat" key={label}>
                  <span>{label}</span>
                  <strong>{value}</strong>
                </div>
              ))}
            </div>
          )}

          {errorsQuery.data?.data.errors && errorsQuery.data.data.errors.length > 0 && (
            <div>
              <p style={{ fontSize: '0.85rem', fontWeight: 600, marginBottom: '0.4rem' }}>
                Строки с ошибками:
              </p>
              <div className="import-errors">
                {errorsQuery.data.data.errors.map((e, i) => (
                  <div className="import-error" key={i}>
                    Строка {e.row}: {e.message}
                  </div>
                ))}
              </div>
            </div>
          )}

          {job.status === 'completed' && (
            <Notice tone="success">Импорт успешно завершён.</Notice>
          )}
        </div>
      )}
    </Container>
  );
}
