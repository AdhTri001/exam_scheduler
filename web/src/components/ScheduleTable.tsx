import React, { useState, useMemo } from 'react';
import { Paper, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, TableSortLabel, TextField, Box, Typography } from '@mui/material';

interface ScheduleTableProps {
  scheduleData: any[];
}

type Order = 'asc' | 'desc';

export const ScheduleTable: React.FC<ScheduleTableProps> = ({ scheduleData }) => {
  const [filter, setFilter] = useState('');
  const [orderBy, setOrderBy] = useState('course_id');
  const [order, setOrder] = useState<Order>('asc');

  const handleSort = (property: string) => {
    const isAsc = orderBy === property && order === 'asc';
    setOrder(isAsc ? 'desc' : 'asc');
    setOrderBy(property);
  };

  const sortedAndFilteredData = useMemo(() => {
    const stabilizedB = scheduleData.map((el, index) => [el, index] as [any, number]);
    stabilizedB.sort((a, b) => {
      const orderVal = order === 'asc' ? 1 : -1;
      if (a[0][orderBy] < b[0][orderBy]) return -1 * orderVal;
      if (a[0][orderBy] > b[0][orderBy]) return 1 * orderVal;
      return a[1] - b[1];
    });

    return stabilizedB.map(el => el[0]).filter(row =>
      Object.values(row).some(value =>
        String(value).toLowerCase().includes(filter.toLowerCase())
      )
    );
  }, [scheduleData, order, orderBy, filter]);

  if (scheduleData.length === 0) return null;

  const headers = Object.keys(scheduleData[0] || {});

  return (
    <Paper sx={{ mt: 2, p: 2 }}>
        <Typography variant="h6" gutterBottom>Generated Schedule</Typography>
      <Box sx={{ mb: 2 }}>
        <TextField
          label="Filter schedule..."
          variant="outlined"
          fullWidth
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
        />
      </Box>
      <TableContainer>
        <Table stickyHeader>
          <TableHead>
            <TableRow>
              {headers.map(header => (
                <TableCell key={header} sortDirection={orderBy === header ? order : false}>
                  <TableSortLabel
                    active={orderBy === header}
                    direction={orderBy === header ? order : 'asc'}
                    onClick={() => handleSort(header)}
                  >
                    {header.replace(/_/g, ' ')}
                  </TableSortLabel>
                </TableCell>
              ))}
            </TableRow>
          </TableHead>
          <TableBody>
            {sortedAndFilteredData.map((row, index) => (
              <TableRow key={index} hover>
                {headers.map(header => (
                  <TableCell key={header}>{row[header]}</TableCell>
                ))}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
    </Paper>
  );
};
