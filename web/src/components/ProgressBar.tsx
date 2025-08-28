import React from 'react';
import { Box, LinearProgress, Typography } from '@mui/material';

interface ProgressBarProps {
  message: string;
  value?: number;
}

export const ProgressBar: React.FC<ProgressBarProps> = ({ message, value }) => {
  return (
    <Box sx={{ width: '100%', my: 2 }}>
      <Typography variant="body1" gutterBottom>{message}</Typography>
      {value !== undefined ? (
        <LinearProgress variant="determinate" value={value} />
      ) : (
        <LinearProgress />
      )}
    </Box>
  );
};
