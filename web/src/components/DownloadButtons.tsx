import React from 'react';
import { Button, Box, Typography, Card, CardContent } from '@mui/material';
import { unparseCsv } from '../lib/csv';
import type { SuccessResponse } from '../lib/wasmTypes';

interface DownloadButtonsProps {
  scheduleData: any[];
  result: SuccessResponse;
}

export const DownloadButtons: React.FC<DownloadButtonsProps> = ({ scheduleData, result }) => {
  const handleDownload = (filename: string, content: string, mimeType: string) => {
    const blob = new Blob([content], { type: mimeType });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  const handleDownloadJson = () => {
    const jsonData = {
      schedule: scheduleData,
      report: result.report,
      stats: result.stats,
      generatedAt: new Date().toISOString(),
    };
    handleDownload('schedule-report.json', JSON.stringify(jsonData, null, 2), 'application/json');
  };

  const handleDownloadCsv = () => {
    // Prefer the raw CSV from wasm, but fall back to unparsing if needed
    const csvContent = result.scheduleCSV || unparseCsv(scheduleData);
    handleDownload('schedule.csv', csvContent, 'text/csv;charset=utf-8;');
  };

  return (
    <Card elevation={1} sx={{ mt: 2, bgcolor: 'background.paper' }}>
      <CardContent>
        <Typography variant="h6" gutterBottom sx={{ color: 'primary.main', fontWeight: 600 }}>
          📥 Download Results
        </Typography>
        <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
          Export your schedule and validation report
        </Typography>
        <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap' }}>
          <Button
            variant="contained"
            onClick={handleDownloadCsv}
            sx={{ minWidth: 180 }}
          >
            📄 Download CSV
          </Button>
          <Button
            variant="outlined"
            onClick={handleDownloadJson}
            sx={{ minWidth: 180 }}
          >
            📊 Download JSON Report
          </Button>
        </Box>
      </CardContent>
    </Card>
  );
};
