import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { Container, Navbar, Nav } from 'react-bootstrap';
import 'bootstrap/dist/css/bootstrap.min.css';
import './App.css';

import Home from './pages/Home';
import ImageDetails from './components/ImageDetails';
import TaggedImages from './components/TaggedImages';
import TagsList from './components/TagsList';

function App() {
  return (
    <Router>
      <div className="App">
        <Navbar bg="dark" variant="dark" expand="lg">
          <Container>
            <Navbar.Brand href="/">MinIO Նկարների Պատկերասրահ</Navbar.Brand>
            <Navbar.Toggle aria-controls="basic-navbar-nav" />
            <Navbar.Collapse id="basic-navbar-nav">
              <Nav className="ms-auto">
                <Nav.Link href="/">Գլխավոր</Nav.Link>
                <Nav.Link href="/tags">Թեգեր</Nav.Link>
              </Nav>
            </Navbar.Collapse>
          </Container>
        </Navbar>

        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/images/:id" element={<ImageDetails />} />
          <Route path="/tags" element={<TagsList />} />
          <Route path="/tags/:tag" element={<TaggedImages />} />
        </Routes>

        <footer className="footer bg-light py-3 mt-auto">
          <Container className="text-center">
            <p className="mb-0">MinIO Պատկերասրահ &copy; {new Date().getFullYear()}</p>
          </Container>
        </footer>
      </div>
    </Router>
  );
}

export default App;