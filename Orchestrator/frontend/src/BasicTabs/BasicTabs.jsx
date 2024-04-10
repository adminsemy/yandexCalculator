import * as React from 'react';
import Tabs from '@mui/material/Tabs';
import Tab from '@mui/material/Tab';
import Box from '@mui/material/Box';
import CustomTabPanel from '../CustomTabPanel/CustomTabPanel';
import TwoTabPanel from '../TwoTabPanel/TwoTabPanel';
import ThreeTabPanel from '../ThreeTabPanel/ThreeTabPanel';
import Grid from '@mui/material/Grid';
import Button from '@mui/material/Button';
import Tooltip from '@mui/material/Tooltip';
  
  export default function BasicTabs({client}) {
    const [value, setValue] = React.useState(0);
  
    const handleChange = (event, newValue) => {
      setValue(newValue);
    };
  
    return (
      <Box sx={{ width: '100%' }}>
        <Grid container justifyContent="center">
          <Grid item container xs={12} alignItems="flex-end" direction="column">
            <Grid item>
              <Tooltip title="Add" placement="right-start">
                <Button>Выход</Button>
              </Tooltip>
            </Grid>
          </Grid>
        </Grid>
        <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
          <Tabs value={value} onChange={handleChange} aria-label="basic tabs example">
            <Tab label="Рассчитать выражение" />
            <Tab label="Установить продолжительность" />
            <Tab label="Мониторинг рабочих агентов"  />
          </Tabs>
        </Box>
        <CustomTabPanel value={value} index={0} client={client}/>
        <TwoTabPanel value={value} index={1} client={client} />
        <ThreeTabPanel value={value} index={2} client={client} />
      </Box>
    );
  }