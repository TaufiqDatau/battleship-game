import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-board',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './board.component.html',
  styleUrls: ['./board.component.scss'],
})
export class BoardComponent {
  board: string[][] = [];
  socket = new WebSocket('http://localhost:8080/ws');

  ngOnInit() {
    this.initializeBoard();
    console.log('Attempting Connection...');

    this.socket.onopen = () => {
      console.log('Successfully Connected');
      this.socket.send('Hi From the Client!');
    };

    this.socket.onclose = (event) => {
      console.log('Socket Closed Connection: ', event);
      this.socket.send('Client Closed!');
    };

    this.socket.onerror = (error) => {
      console.log('Socket Error: ', error);
    };

    this.socket.onmessage = (event) => {
      try {
        const receivedData = JSON.parse(event.data);

        if (receivedData.condition === 'miss') {
          this.board[receivedData.row][receivedData.col] = 'miss';
        } else {
          this.board[receivedData.row][receivedData.col] = 'hit';
        }
      } catch (error) {
        console.error('Error parsing JSON:', error);
      }
    };
  }

  initializeBoard() {
    this.board = Array(10)
      .fill(null)
      .map(() => Array(10).fill('empty'));
  }

  fireShot(row: number, col: number) {
    const jsonData = {
      action: 'fire_shot',
      row: row,
      col: col,
    };

    const jsonString = JSON.stringify(jsonData);
    this.socket.send(jsonString);
  }

  getCellClass(cell: string) {
    return {
      'empty-cell': cell === 'empty',
      'hit-cell': cell === 'hit',
      'miss-cell': cell === 'miss',
    };
  }
}
