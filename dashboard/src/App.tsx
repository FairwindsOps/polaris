import React, { useState, useEffect } from 'react';
import data from './data.json';
import LeftNavBar from './components/Navigation/LeftBar/LeftNavBar';
import TopNavBar from './components/Navigation/TopBar/TopNavBar';
import './App.scss';
// import './App.react.scss';

function App() {
  const [pageDisplay, setPageDisplay] = useState<string>('dashboard');
  const [namespaces, setNamespaces] = useState<string[]>([]);
  const [selectedNamespace, setSelectedNamespace] = useState<string>('');

  useEffect(() => {
    const allNamespaces: string[] = [];
    data.Results.forEach(result => {
      if (!allNamespaces.includes(result.Namespace)) {
        allNamespaces.push(result.Namespace);
      }
    })
    setNamespaces(allNamespaces);
  }, [])

  return (
    <div className="App">
      <LeftNavBar />
      <div className="app-content">
        <TopNavBar pageDisplay={pageDisplay} setPageDisplay={setPageDisplay} namespaces={namespaces} setSelected={setSelectedNamespace} />
        {pageDisplay === 'dashboard' && (
          <h1>Polaris Dashboard</h1>
        )}
        {pageDisplay === 'namespaces' && (
          <h1>{selectedNamespace} Namespace</h1>
        )}
      </div>
    </div>
  );
}

export default App;
