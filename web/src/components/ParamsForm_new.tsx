import React from 'react';
import { TextField, Box, Typography, Card, CardContent, Stack } from '@mui/material';
import type { RunScheduleParams } from '../lib/wasmTypes';

interface ParamsFormProps {
  params: Partial<RunScheduleParams>;
  setParams: (params: Partial<RunScheduleParams>) => void;
}

export const ParamsForm: React.FC<ParamsFormProps> = ({ params, setParams }) => {
  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value, type } = event.target;
    setParams({
      ...params,
      [name]: type === 'number' ? Number(value) : value,
    });
  };

  return (
    <Box sx={{ mt: 2 }}>
      <Typography variant="h5" gutterBottom sx={{ mb: 3, fontWeight: 600 }}>
        📅 Scheduling Configuration
      </Typography>

      <Stack spacing={3}>
        {/* Date Configuration Section */}
        <Card elevation={2}>
          <CardContent>
            <Typography variant="h6" gutterBottom sx={{ color: 'primary.main', fontWeight: 600 }}>
              📅 Exam Period
            </Typography>
            <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: '1fr 1fr' }, gap: 3, mt: 2 }}>
              <TextField
                name="examStartDate"
                label="Start Date"
                type="date"
                value={params.examStartDate || ''}
                onChange={handleChange}
                fullWidth
                InputLabelProps={{ shrink: true }}
                required
                variant="outlined"
                sx={{
                  '& .MuiOutlinedInput-root': {
                    borderRadius: 2,
                    '&:hover': {
                      boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
                    },
                  },
                }}
              />
              <TextField
                name="examEndDate"
                label="End Date"
                type="date"
                value={params.examEndDate || ''}
                onChange={handleChange}
                fullWidth
                InputLabelProps={{ shrink: true }}
                required
                variant="outlined"
                sx={{
                  '& .MuiOutlinedInput-root': {
                    borderRadius: 2,
                    '&:hover': {
                      boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
                    },
                  },
                }}
              />
            </Box>
          </CardContent>
        </Card>

        {/* Schedule Configuration Section */}
        <Card elevation={2}>
          <CardContent>
            <Typography variant="h6" gutterBottom sx={{ color: 'primary.main', fontWeight: 600 }}>
              ⏰ Schedule Settings
            </Typography>
            <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: '1fr 1fr 1fr' }, gap: 3, mt: 2 }}>
              <TextField
                name="slotsPerDay"
                label="Slots Per Day"
                type="number"
                inputProps={{ min: 1, max: 10 }}
                value={params.slotsPerDay || 2}
                onChange={handleChange}
                fullWidth
                variant="outlined"
                helperText="How many exam sessions per day"
                sx={{
                  '& .MuiOutlinedInput-root': {
                    borderRadius: 2,
                    '&:hover': {
                      boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
                    },
                  },
                }}
              />
              <TextField
                name="slotDuration"
                label="Slot Duration (minutes)"
                type="number"
                inputProps={{ min: 30, max: 480, step: 15 }}
                value={params.slotDuration || 180}
                onChange={handleChange}
                fullWidth
                variant="outlined"
                helperText="Duration of each exam session"
                sx={{
                  '& .MuiOutlinedInput-root': {
                    borderRadius: 2,
                    '&:hover': {
                      boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
                    },
                  },
                }}
              />
              <TextField
                name="minGap"
                label="Minimum Gap (minutes)"
                type="number"
                inputProps={{ min: 0, max: 1440, step: 15 }}
                value={params.minGap || 60}
                onChange={handleChange}
                fullWidth
                variant="outlined"
                helperText="Minimum time between exams for same student"
                sx={{
                  '& .MuiOutlinedInput-root': {
                    borderRadius: 2,
                    '&:hover': {
                      boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
                    },
                  },
                }}
              />
            </Box>
          </CardContent>
        </Card>

        {/* Algorithm Configuration Section */}
        <Card elevation={2}>
          <CardContent>
            <Typography variant="h6" gutterBottom sx={{ color: 'primary.main', fontWeight: 600 }}>
              ⚙️ Algorithm Settings
            </Typography>
            <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: '1fr 1fr' }, gap: 3, mt: 2 }}>
              <TextField
                name="tries"
                label="Optimization Attempts"
                type="number"
                inputProps={{ min: 1, max: 1000 }}
                value={params.tries || 100}
                onChange={handleChange}
                fullWidth
                variant="outlined"
                helperText="More attempts = better results (slower)"
                sx={{
                  '& .MuiOutlinedInput-root': {
                    borderRadius: 2,
                    '&:hover': {
                      boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
                    },
                  },
                }}
              />
              <TextField
                name="seed"
                label="Random Seed"
                type="number"
                inputProps={{ min: 0 }}
                value={params.seed || 0}
                onChange={handleChange}
                fullWidth
                variant="oriented"
                helperText="0 = random, same number = reproducible results"
                sx={{
                  '& .MuiOutlinedInput-root': {
                    borderRadius: 2,
                    '&:hover': {
                      boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
                    },
                  },
                }}
              />
            </Box>
          </CardContent>
        </Card>

        {/* Optional Configuration Section */}
        <Card elevation={2}>
          <CardContent>
            <Typography variant="h6" gutterBottom sx={{ color: 'primary.main', fontWeight: 600 }}>
              🔧 Optional Settings
            </Typography>
            <Box sx={{ mt: 2 }}>
              <TextField
                name="timezone"
                label="Timezone"
                value={params.timezone || 'UTC'}
                onChange={handleChange}
                fullWidth
                variant="outlined"
                helperText="IANA timezone (e.g., America/New_York, Europe/London)"
                sx={{
                  '& .MuiOutlinedInput-root': {
                    borderRadius: 2,
                    '&:hover': {
                      boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
                    },
                  },
                }}
              />
            </Box>
          </CardContent>
        </Card>
      </Stack>
    </Box>
  );
};
