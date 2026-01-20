#!/bin/bash

# Tunnel API Microservices Startup Script

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if a port is in use
check_port() {
    local port=$1
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null ; then
        return 0
    else
        return 1
    fi
}

# Function to start a service
start_service() {
    local service_name=$1
    local port=$2
    local cmd=$3
    
    print_status "Starting $service_name on port $port..."
    
    if check_port $port; then
        print_warning "Port $port is already in use. Skipping $service_name."
        return 1
    fi
    
    # Start service in background
    nohup $cmd > logs/${service_name}.log 2>&1 &
    local pid=$!
    echo $pid > pids/${service_name}.pid
    
    # Wait a moment for service to start
    sleep 2
    
    if check_port $port; then
        print_success "$service_name started successfully (PID: $pid)"
        return 0
    else
        print_error "Failed to start $service_name"
        return 1
    fi
}

# Function to stop a service
stop_service() {
    local service_name=$1
    local pid_file="pids/${service_name}.pid"
    
    if [ -f "$pid_file" ]; then
        local pid=$(cat $pid_file)
        print_status "Stopping $service_name (PID: $pid)..."
        
        if kill -0 $pid 2>/dev/null; then
            kill $pid
            sleep 2
            
            if kill -0 $pid 2>/dev/null; then
                print_warning "Service $service_name didn't stop gracefully, force killing..."
                kill -9 $pid
            fi
            
            rm -f $pid_file
            print_success "$service_name stopped"
        else
            print_warning "$service_name is not running"
            rm -f $pid_file
        fi
    else
        print_warning "$service_name is not running (no PID file found)"
    fi
}

# Function to show service status
show_status() {
    echo -e "\n${BLUE}=== Service Status ===${NC}"
    
    local services=("gateway:8080" "billing:8081" "controller:8082" "monitoring:8083")
    
    for service_info in "${services[@]}"; do
        IFS=':' read -r service_name port <<< "$service_info"
        
        if check_port $port; then
            local pid_file="pids/${service_name}.pid"
            if [ -f "$pid_file" ]; then
                local pid=$(cat $pid_file)
                echo -e "${GREEN}✓${NC} $service_name (port $port) - Running (PID: $pid)"
            else
                echo -e "${YELLOW}?${NC} $service_name (port $port) - Port in use but no PID file"
            fi
        else
            echo -e "${RED}✗${NC} $service_name (port $port) - Not running"
        fi
    done
    echo ""
}

# Function to show logs
show_logs() {
    local service_name=$1
    local log_file="logs/${service_name}.log"
    
    if [ -f "$log_file" ]; then
        echo -e "\n${BLUE}=== $service_name Logs ===${NC}"
        tail -f "$log_file"
    else
        print_error "Log file for $service_name not found"
    fi
}

# Create necessary directories
mkdir -p logs pids

# Main script logic
case "${1:-start}" in
    "start")
        print_status "Starting Tunnel API Microservices..."
        
        # Start services in order
        start_service "billing" "8081" "go run cmd/billing/main.go" || true
        start_service "controller" "8082" "go run cmd/controller/main.go" || true
        start_service "monitoring" "8083" "go run cmd/monitoring/main.go" || true
        start_service "gateway" "8080" "go run cmd/gateway/main.go" || true
        
        sleep 3
        show_status
        
        print_success "All services started. API Gateway available at http://localhost:8080"
        print_status "Use './scripts/start-services.sh status' to check service status"
        print_status "Use './scripts/start-services.sh logs <service>' to view logs"
        ;;
    
    "stop")
        print_status "Stopping Tunnel API Microservices..."
        
        stop_service "gateway"
        stop_service "monitoring"
        stop_service "controller"
        stop_service "billing"
        
        print_success "All services stopped"
        ;;
    
    "restart")
        print_status "Restarting Tunnel API Microservices..."
        $0 stop
        sleep 2
        $0 start
        ;;
    
    "status")
        show_status
        ;;
    
    "logs")
        if [ -z "$2" ]; then
            print_error "Please specify a service name (gateway, billing, controller, monitoring)"
            exit 1
        fi
        show_logs "$2"
        ;;
    
    "docker")
        print_status "Starting services with Docker Compose..."
        docker-compose up -d
        print_success "Services started with Docker Compose"
        ;;
    
    "docker-stop")
        print_status "Stopping Docker Compose services..."
        docker-compose down
        print_success "Docker Compose services stopped"
        ;;
    
    *)
        echo "Usage: $0 {start|stop|restart|status|logs <service>|docker|docker-stop}"
        echo ""
        echo "Commands:"
        echo "  start       - Start all microservices"
        echo "  stop        - Stop all microservices"
        echo "  restart     - Restart all microservices"
        echo "  status      - Show status of all services"
        echo "  logs <svc>  - Show logs for specific service"
        echo "  docker      - Start services with Docker Compose"
        echo "  docker-stop - Stop Docker Compose services"
        echo ""
        echo "Services: gateway, billing, controller, monitoring"
        exit 1
        ;;
esac 