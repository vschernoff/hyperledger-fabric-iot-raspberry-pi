import React from 'react';
import {ComposedChart, ResponsiveContainer, Line, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend} from 'recharts';

class LineBarAreaComposedChart extends React.Component {
  render() {
    const {elements, lines, keyChart} = this.props;

    if (!elements) {
      return null;
    }

    const data = elements.map((r) => {
      return {
        ...r,
        ts: (new Date(r.value.timestamp * 1000)).toLocaleString()
      };
    })
      .sort((a, b) => {
        return a.value.timestamp > b.value.timestamp ? 1 : -1;
      });

    return !!data.length && (
      <div>
        <div className={`container`}>
          <h3>{keyChart}</h3>
        </div>
        <ResponsiveContainer width="100%" height={600}>
          <ComposedChart data={data}
                         margin={{top: 20, right: 20, bottom: 20, left: 20}}>
            <CartesianGrid strokeDasharray="3 3"/>
            <XAxis dataKey="ts" name="Time"/>
            <YAxis/>
            <Tooltip/>
            <Legend/>
            {
              lines.map((line, index) => {
                return (<Line name={line.name} dataKey={line.key} stroke={line.color} key={index}/>)
              })
            }
          </ComposedChart>
        </ResponsiveContainer>
      </div>
    );
  }
}

export {LineBarAreaComposedChart as Chart};
