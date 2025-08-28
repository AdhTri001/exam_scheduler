import React from 'react';
import { Paper, Typography, Box, Alert, List, ListItem, ListItemText, Chip } from '@mui/material';
import type { ValidationReport } from '../lib/wasmTypes';

interface ValidationPanelProps {
  report: ValidationReport;
}

export const ValidationPanel: React.FC<ValidationPanelProps> = ({ report }) => {
  return (
    <Paper sx={{ p: 2, mt: 2 }}>
      <Typography variant="h6">Validation Report</Typography>
      <Box sx={{ mt: 2 }}>
        {report.valid ? (
          <Alert severity="success">Schedule is valid!</Alert>
        ) : (
          <Alert severity="error">Schedule has issues that need attention.</Alert>
        )}

        <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1, mt: 2 }}>
            <Chip label={`Conflicts: ${report.conflicts || 0}`} color={(report.conflicts || 0) > 0 ? 'error' : 'success'} />
            <Chip label={`Unassigned Courses: ${report.unassigned?.length || 0}`} color={(report.unassigned?.length || 0) > 0 ? 'warning' : 'default'} />
        </Box>

        {report.studentClashes && report.studentClashes.length > 0 && (
             <Box sx={{ mt: 2 }}>
                <Typography variant="subtitle1">Student Clashes:</Typography>
                <Paper variant="outlined" sx={{ maxHeight: 200, overflow: 'auto', p: 1 }}>
                    <List dense>
                        {report.studentClashes.map((clash, index) => (
                            <ListItem key={index}><ListItemText primary={clash} /></ListItem>
                        ))}
                    </List>
                </Paper>
            </Box>
        )}

        {report.unassigned && report.unassigned.length > 0 && (
            <Box sx={{ mt: 2 }}>
                <Typography variant="subtitle1">Unassigned Courses:</Typography>
                 <Paper variant="outlined" sx={{ maxHeight: 200, overflow: 'auto', p: 1 }}>
                    <List dense>
                        {report.unassigned.map((course, index) => (
                            <ListItem key={index}><ListItemText primary={course} /></ListItem>
                        ))}
                    </List>
                </Paper>
            </Box>
        )}

        {report.capacityWarnings && report.capacityWarnings.length > 0 && (
          <Box sx={{ mt: 2 }}>
            <Typography variant="subtitle1">Capacity Warnings:</Typography>
            <Paper variant="outlined" sx={{ maxHeight: 200, overflow: 'auto', p: 1 }}>
                <List dense>
                {report.capacityWarnings.map((warning, index) => (
                    <ListItem key={index}>
                    <ListItemText primary={warning} />
                    </ListItem>
                ))}
                </List>
            </Paper>
          </Box>
        )}

        {report.errors && report.errors.length > 0 && (
             <Box sx={{ mt: 2 }}>
                <Typography variant="subtitle1">Fatal Errors:</Typography>
                <Paper variant="outlined" sx={{ maxHeight: 200, overflow: 'auto', p: 1 }}>
                    <List dense>
                    {report.errors.map((error, index) => (
                        <ListItem key={index}>
                        <ListItemText primary={error} />
                        </ListItem>
                    ))}
                    </List>
                </Paper>
            </Box>
        )}
      </Box>
    </Paper>
  );
};
