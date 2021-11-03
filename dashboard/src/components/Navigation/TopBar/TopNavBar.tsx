import React, { SyntheticEvent, useState } from 'react';
import { Nav, Navbar, NavDropdown } from 'react-bootstrap';
import './TopNavBar.scss';

type NavProps = {
  pageDisplay: string;
  setPageDisplay: React.Dispatch<React.SetStateAction<string>>;
  namespaces: string[];
  setSelected: React.Dispatch<React.SetStateAction<string>>;
};

const TopNavBar = ({
  pageDisplay,
  setPageDisplay,
  namespaces,
  setSelected,
}: NavProps): JSX.Element => {
  const [selectedNamespace, setSelectedNamespace] =
    useState<string>('All Namespaces');

  const handleNamespaceSelection = (
    namespace: SyntheticEvent<HTMLDivElement, Event>
  ): void => {
    if (namespace.toString() === 'All Namespaces') {
      setSelectedNamespace('All Namespace');
    } else if (namespace) {
      setSelectedNamespace(namespace.toString());
    }
    setPageDisplay('namespaces');
    setSelected(namespace.toString());
  };

  const NavItems = (): JSX.Element => (
    <>
      {namespaces.map((item) => {
        return (
          <NavDropdown.Item key={item} eventKey={item}>
            {item}
          </NavDropdown.Item>
        );
      })}
    </>
  );

  return (
    <header>
      <Navbar className="top-nav-bar" expand={true}>
        <Nav variant="tabs" className="top-nav">
          <NavDropdown
            title={selectedNamespace}
            id="nav-dropdown"
            active={pageDisplay === 'namespaces'}
            onSelect={handleNamespaceSelection}
          >
            <NavDropdown.Item
              key={'all-namespaces'}
              eventKey={'all-namespaces'}
            >
              All Namespaces
            </NavDropdown.Item>
            <NavItems />
          </NavDropdown>
          <Nav.Item onClick={() => setPageDisplay('dashboard')}>
            <Nav.Link active={pageDisplay === 'dashboard'}>Dashboard</Nav.Link>
          </Nav.Item>
        </Nav>
      </Navbar>
    </header>
  );
};

export default TopNavBar;
